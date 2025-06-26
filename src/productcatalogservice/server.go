// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/golang/protobuf/jsonpb"
	"github.com/sirupsen/logrus"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"

	grpcotel "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	cat          pb.ListProductsResponse
	catalogMutex *sync.Mutex
	log          *logrus.Logger
	extraLatency time.Duration

	port = "3550"

	reloadCatalog bool
	db            *sql.DB
)

func main() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
	initTracer(log)

	err := initDB() // 调用输出化数据库的函数
	if err != nil {
		fmt.Printf("init db failed,err:%v\n", err)
		return
	}

	catalogMutex = &sync.Mutex{}
	// err = readCatalogFile(&cat)
	err = readCatalogDB(&cat)
	if err != nil {
		log.Warnf("could not parse product catalog")
	}
	if os.Getenv("DISABLE_TRACING") == "" {
		log.Info("Tracing enabled.")

	} else {
		log.Info("Tracing disabled.")
	}

	flag.Parse()

	// set injected latency
	if s := os.Getenv("EXTRA_LATENCY"); s != "" {
		v, err := time.ParseDuration(s)
		if err != nil {
			log.Fatalf("failed to parse EXTRA_LATENCY (%s) as time.Duration: %+v", v, err)
		}
		extraLatency = v
		log.Infof("extra latency enabled (duration: %v)", extraLatency)
	} else {
		extraLatency = time.Duration(0)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for {
			sig := <-sigs
			log.Printf("Received signal: %s", sig)
			if sig == syscall.SIGUSR1 {
				reloadCatalog = true
				log.Infof("Enable catalog reloading")
			} else {
				reloadCatalog = false
				log.Infof("Disable catalog reloading")
			}
		}
	}()

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	log.Infof("starting grpc server at :%s", port)
	run(port)
	select {}
}

func run(port string) string {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}
	var srv *grpc.Server
	if os.Getenv("DISABLE_STATS") == "" {
		log.Info("Stats enabled.")
		srv = grpc.NewServer(
			grpc.UnaryInterceptor(grpcotel.UnaryServerInterceptor(global.Tracer(""))),
			grpc.StreamInterceptor(grpcotel.StreamServerInterceptor(global.Tracer(""))),
		)
	} else {
		log.Info("Stats disabled.")
		srv = grpc.NewServer()
	}

	svc := &productCatalog{}

	pb.RegisterProductCatalogServiceServer(srv, svc)
	healthpb.RegisterHealthServer(srv, svc)
	go srv.Serve(l)
	return l.Addr().String()
}

func initTracer(log logrus.FieldLogger) func() {
	podIp := os.Getenv("POD_IP")
	podName := os.Getenv("POD_NAME")
	nodeName := os.Getenv("NODE_NAME")
	svcAddr := os.Getenv("JAEGER_SERVICE_ADDR")
	serviceName := os.Getenv("SERVICE_NAME")
	namespace := os.Getenv("NAMESPACE")
	if svcAddr == "" {
		log.Info("jaeger initialization disabled.")
	}
	endPoint := fmt.Sprintf("http://%s", svcAddr)
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint(endPoint),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: serviceName,
			Tags: []kv.KeyValue{
				kv.String("exporter", "jaeger"),
				kv.String("ip", podIp),
				kv.String("name", podName),
				kv.String("node_name", nodeName),
				kv.String("namespace", namespace),
			},
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("jaeger initialization completed.")
	return func() {
		flush()
	}
}

type productCatalog struct{}

func readCatalogFile(catalog *pb.ListProductsResponse) error {
	catalogMutex.Lock()
	defer catalogMutex.Unlock()
	catalogJSON, err := ioutil.ReadFile("products.json")
	if err != nil {
		log.Fatalf("failed to open product catalog json file: %v", err)
		return err
	}
	if err := jsonpb.Unmarshal(bytes.NewReader(catalogJSON), catalog); err != nil {
		log.Warnf("failed to parse the catalog JSON: %v", err)
		return err
	}
	log.Info("successfully parsed product catalog json")
	return nil
}

func initDB() (err error) {
	mysqlAddr := os.Getenv("MYSQL_ADDR")
	connMaxLifeTime, _ := strconv.Atoi(os.Getenv("ConnMaxLifeTime"))
	maxIdleConns, _ := strconv.Atoi(os.Getenv("mySQLmaxIdleConns"))

	// connect to sql
	user := os.Getenv("SQL_USER")
	pwd := os.Getenv("SQL_PASSWORD")
	// user := "root"
	// pwd := "root123"
	path := strings.Join([]string{user, ":", pwd, "@tcp(", mysqlAddr, ")/productdb"}, "")
	db, err = sql.Open("mysql", path)
	if err = db.Ping(); err != nil {
		log.Fatalf("failed to open mysql database: %v", err)
		return err
	}
	db.SetConnMaxLifetime(time.Duration(connMaxLifeTime))
	db.SetMaxIdleConns(maxIdleConns)

	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping mysql database: %v", err)
		return err
	}
	return nil
}

type Product struct {
	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Picture     string `protobuf:"bytes,4,opt,name=picture,proto3" json:"picture,omitempty"`
	PriceUsd    string `protobuf:"bytes,5,opt,name=price_usd,json=priceUsd,proto3" json:"price_usd,omitempty"`
	Categories  string `protobuf:"bytes,6,rep,name=categories,proto3" json:"categories,omitempty"`
}

func readCatalogDB(catalog *pb.ListProductsResponse) error {
	catalogMutex.Lock()
	defer catalogMutex.Unlock()

	var productList []*pb.Product
	// rows, err := db.Query("SELECT a.product_id, a.item_name, a.description, a.picture_path, a.categorise_list, b.currencyCode, b.units, b.nanos FROM products a INNER JOIN price b ON a.product_id = b.product_id")
	rows, err := db.Query("SELECT id AS product_id, name AS item_name, description, picture AS picture_path, categories AS categorise_list, JSON_UNQUOTE(JSON_EXTRACT(priceUsd, '$.currencyCode')) AS currencyCode, JSON_UNQUOTE(JSON_EXTRACT(priceUsd, '$.units')) AS units, JSON_UNQUOTE(JSON_EXTRACT(priceUsd, '$.nanos')) AS nanos FROM products;")
	if err != nil {
		log.Warnf("failed to query mysql: %v", err)
		return err
	}
	// Parse query results to Product struct
	for rows.Next() {
		var product pb.Product
		var money pb.Money
		var categoriesStr string
		err := rows.Scan(&product.Id, &product.Name, &product.Description, &product.Picture, &categoriesStr, &money.CurrencyCode, &money.Units, &money.Nanos)
		if err != nil {
			log.Warnf("failed to scan mysql query result: %v", err)
			return err
		}
		product.Categories = strings.Split(categoriesStr, ";")
		product.PriceUsd = &money
		productList = append(productList, &product)
		log.Infof("mysql read results: %v", product)
	}
	defer rows.Close()
	catalog.Products = productList

	// log.Infof("successfully parsed product catalog json: %v", catalog.Products)
	return nil
}

func parseCatalog() []*pb.Product {
	if reloadCatalog || len(cat.Products) == 0 {
		err := readCatalogFile(&cat)
		if err != nil {
			return []*pb.Product{}
		}
	}
	return cat.Products
}

func (p *productCatalog) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (p *productCatalog) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

func (p *productCatalog) ListProducts(context.Context, *pb.Empty) (*pb.ListProductsResponse, error) {
	time.Sleep(extraLatency)
	return &pb.ListProductsResponse{Products: parseCatalog()}, nil
}

func (p *productCatalog) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	time.Sleep(extraLatency)
	var found pb.Product
	var money pb.Money
	var categoriesStr string

	// querySQL := "SELECT a.product_id, a.item_name, a.description, a.picture_path, a.categorise_list, b.currencyCode, b.units, b.nanos FROM products a INNER JOIN price b ON a.product_id = b.product_id WHERE a.product_id='" + req.Id + "'"
	querySQL := "SELECT " +
    "id AS product_id, " +
    "name AS item_name, " +
    "description, " +
    "picture AS picture_path, " +
    "categories AS categorise_list, " +
    "JSON_UNQUOTE(JSON_EXTRACT(priceUsd, '$.currencyCode')) AS currencyCode, " +
    "JSON_UNQUOTE(JSON_EXTRACT(priceUsd, '$.units')) AS units, " +
    "JSON_UNQUOTE(JSON_EXTRACT(priceUsd, '$.nanos')) AS nanos " +
    "FROM products " +
    "WHERE id = '" + req.Id + "'"

	rows, err := db.Query(querySQL)
	if err != nil {
		log.Warnf("failed to query product by id: %v", err)
		return nil, err
	}
	for rows.Next() {
		err := rows.Scan(&found.Id, &found.Name, &found.Description, &found.Picture, &categoriesStr, &money.CurrencyCode, &money.Units, &money.Nanos)
		if err != nil {
			log.Warnf("failed to scan mysql query result: %v", err)
			return nil, err
		}
		found.Categories = strings.Split(categoriesStr, ";")
		found.PriceUsd = &money
	}
	defer rows.Close()
	if &found == nil {
		return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)
	}
	return &found, nil
}

func (p *productCatalog) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	time.Sleep(extraLatency)
	// Intepret query as a substring match in name or description.
	var ps []*pb.Product
	// querySQL := "SELECT a.product_id, a.item_name, a.description, a.picture_path, a.categorise_list, b.currencyCode, b.units, b.nanos FROM products a INNER JOIN price b ON a.product_id = b.product_id WHERE item_name='" + req.Query + "'"
	querySQL := "SELECT " +
    "id AS product_id, " +
    "name AS item_name, " +
    "description, " +
    "picture AS picture_path, " +
    "categories AS categorise_list, " +
    "JSON_UNQUOTE(JSON_EXTRACT(priceUsd, '$.currencyCode')) AS currencyCode, " +
    "JSON_UNQUOTE(JSON_EXTRACT(priceUsd, '$.units')) AS units, " +
    "JSON_UNQUOTE(JSON_EXTRACT(priceUsd, '$.nanos')) AS nanos " +
    "FROM products " +
    "WHERE name = '" + req.Query + "'"

	rows, err := db.Query(querySQL)
	if err != nil {
		log.Warnf("failed to query product by name: %v", err)
		return nil, err
	}
	for rows.Next() {
		var found pb.Product
		var money pb.Money
		var categoriesStr string
		err := rows.Scan(&found.Id, &found.Name, &found.Description, &found.Picture, &categoriesStr, &money.CurrencyCode, &money.Units, &money.Nanos)
		if err != nil {
			log.Warnf("failed to scan mysql query result: %v", err)
			return nil, err
		}
		found.Categories = strings.Split(categoriesStr, ";")
		found.PriceUsd = &money
		ps = append(ps, &found)
	}
	defer rows.Close()
	return &pb.SearchProductsResponse{Results: ps}, nil
}
