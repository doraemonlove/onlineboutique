/*
 * Copyright 2018, Google LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package hipstershop;

import com.google.common.collect.ImmutableListMultimap;
import com.google.common.collect.ImmutableMap;
import com.google.common.collect.Iterables;
import hipstershop.Demo.Ad;
import hipstershop.Demo.AdRequest;
import hipstershop.Demo.AdResponse;
import io.grpc.Context;
import io.grpc.Contexts;
import io.grpc.Grpc;
import io.grpc.Metadata;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.ServerCall;
import io.grpc.ServerCallHandler;
import io.grpc.stub.StreamObserver;
import io.grpc.StatusRuntimeException;
import io.grpc.health.v1.HealthCheckResponse.ServingStatus;
import io.grpc.services.*;
import io.grpc.stub.StreamObserver;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.Random;
import java.util.HashMap;
import java.net.InetSocketAddress;
import java.util.concurrent.TimeUnit;
import org.apache.logging.log4j.Level;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import io.opentelemetry.OpenTelemetry;
import io.opentelemetry.exporters.zipkin.ZipkinSpanExporter;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.trace.export.SimpleSpanProcessor;
import io.opentelemetry.trace.Span;
import io.opentelemetry.trace.Tracer;
import io.opentelemetry.context.ContextUtils;
import io.opentelemetry.context.Scope;
import io.opentelemetry.context.propagation.HttpTextFormat;
import io.opentelemetry.sdk.trace.TracerSdkProvider;
import io.opentelemetry.trace.Tracer;
import io.opentelemetry.trace.Status;
import java.sql.*;
import java.lang.ClassNotFoundException;

public final class AdService {

  private static final Logger logger = LogManager.getLogger(AdService.class);
  private Tracer tracer = OpenTelemetry.getTracer("AdService");

  private static final String podName = System.getenv("POD_NAME");
  private static final String nodeName = System.getenv("NODE_NAME");
  private static final String namespace = System.getenv("NAMESPACE");

  private static final String ENDPOINT_V2_SPANS = "/api/v2/spans";
  private static final String ip = System.getenv("JAEGER_HOST");
  private static final String port = System.getenv("ZIPKIN_PORT");
  private static final String serviceName = System.getenv("SERVICE_NAME");
  
  private static final String mysqlAddr = System.getenv("MYSQL_ADDR");
  private static final String user = System.getenv("SQL_USER");
  private static final String pwd = System.getenv("SQL_PASSWORD");

  private static final String endpoint = String.format("http://%s:%s%s", ip, port, ENDPOINT_V2_SPANS );
  private ZipkinSpanExporter exporter  =  ZipkinSpanExporter.newBuilder()
                                                    .setEndpoint(endpoint)
                                                    .setServiceName(serviceName)
                                                    .build();

  private HttpTextFormat textFormat = OpenTelemetry.getPropagators().getHttpTextFormat();

  // Extract the Distributed Context from the gRPC metadata
  HttpTextFormat.Getter<Metadata> getter =
      new HttpTextFormat.Getter<Metadata>() {
        @Override
        public String get(Metadata carrier, String key) {
          Metadata.Key<String> k = Metadata.Key.of(key, Metadata.ASCII_STRING_MARSHALLER);
          if (carrier.containsKey(k)) {
            return carrier.get(k);
          }
          return "";
        }
      };


  @SuppressWarnings("FieldCanBeLocal")
  private static int MAX_ADS_TO_SERVE = 2;

  private Server server;
  private HealthStatusManager healthMgr;

  private static final AdService service = new AdService();

  private void start() throws IOException {
    int port = Integer.parseInt(System.getenv().getOrDefault("PORT", "9555"));
    healthMgr = new HealthStatusManager();

    server = ServerBuilder.forPort(port)
              .addService(new AdServiceImpl())
              .addService(healthMgr.getHealthService())
              .intercept(new OpenTelemetryServerInterceptor())
              .build()
              .start();
    logger.info("Ad Service started, listening on " + port);
    Runtime.getRuntime()
        .addShutdownHook(
            new Thread(
                () -> {
                  // Use stderr here since the logger may have been reset by its JVM shutdown hook.
                  System.err.println(
                      "*** shutting down gRPC ads server since JVM is shutting down");
                  AdService.this.stop();
                  System.err.println("*** server shut down");
                }));
    healthMgr.setStatus("", ServingStatus.SERVING);
      
    OpenTelemetrySdk.getTracerProvider()
                .addSpanProcessor(SimpleSpanProcessor.newBuilder(exporter).build());
  }

  private void stop() {
    if (server != null) {
      healthMgr.clearStatus("");
      server.shutdown();
    }
  }

  private static class AdServiceImpl extends hipstershop.AdServiceGrpc.AdServiceImplBase {

    /**
     * Retrieves ads based on context provided in the request {@code AdRequest}.
     *
     * @param req the request containing context.
     * @param responseObserver the stream observer which gets notified with the value of {@code
     *     AdResponse}
     */
    @Override
    public void getAds(AdRequest req, StreamObserver<AdResponse> responseObserver) {
      AdService service = AdService.getInstance();
      try {
        List<Ad> allAds = new ArrayList<>();
        logger.info("received ad request (context_words=" + req.getContextKeysList() + ")");
        if (req.getContextKeysCount() > 0) {
          for (int i = 0; i < req.getContextKeysCount(); i++) {
            Collection<Ad> ads = service.getAdsByCategory(req.getContextKeys(i));
            allAds.addAll(ads);
          }
        } else {
          //span.addAnnotation("No Context provided. Constructing random Ads.");
          allAds = service.getRandomAds();
        }
        if (allAds.isEmpty()) {
          // Serve random ads.
          //span.addAnnotation("No Ads found based on context. Constructing random Ads.");
          allAds = service.getRandomAds();
        }
        AdResponse reply = AdResponse.newBuilder().addAllAds(allAds).build();
        responseObserver.onNext(reply);
        responseObserver.onCompleted();
      } catch (StatusRuntimeException e) {
        logger.log(Level.WARN, "GetAds Failed with status {}", e.getStatus());
        responseObserver.onError(e);
      }
    }
  }

  private static final ImmutableListMultimap<String, Ad> adsMap = createAdsMap();

  private Collection<Ad> getAdsByCategory(String category) {
    return adsMap.get(category);
  }

  private static final Random random = new Random();

  private List<Ad> getRandomAds() {
    List<Ad> ads = new ArrayList<>(MAX_ADS_TO_SERVE);
    Collection<Ad> allAds = adsMap.values();
    for (int i = 0; i < MAX_ADS_TO_SERVE; i++) {
      ads.add(Iterables.get(allAds, random.nextInt(allAds.size())));
    }
    return ads;
  }

  private static AdService getInstance() {
    return service;
  }

  /** Await termination on the main thread since the grpc library uses daemon threads. */
  private void blockUntilShutdown() throws InterruptedException {
    if (server != null) {
      server.awaitTermination();
    }
  }

  private static ImmutableListMultimap<String, Ad> createAdsMap() {
    HashMap<String, Ad> aditemMap = new HashMap<String, Ad>();
    // connect to mysql
    String JDBC_DRIVER = "com.mysql.cj.jdbc.Driver";
    String DB_URL = "jdbc:mysql://" + mysqlAddr + "/addatabase?useSSL=false&allowPublicKeyRetrieval=true&serverTimezone=UTC";

    // set user and pwd
    String USER = user;
    String PASS = pwd;
        
    Connection conn = null;
    Statement stmt = null;

    try {
      // open link
      Class.forName(JDBC_DRIVER);
      conn = DriverManager.getConnection(DB_URL, USER, PASS);

      // execute query
      stmt = conn.createStatement();
      String sql;
      sql = "SELECT * FROM aditems";
      ResultSet rs = stmt.executeQuery(sql);

      // spread result set
      while (rs.next()) {
        String item_name = rs.getString("item_name");
        String redirecturl = rs.getString("redirect_url");
        String text = rs.getString("text");

        aditemMap.put(item_name, Ad.newBuilder()
                                    .setRedirectUrl(redirecturl)
                                    .setText(text)
                                    .build());
      }
      rs.close();
      stmt.close();
      conn.close();
    } catch (ClassNotFoundException e) {
      logger.fatal("MYSQL JDBC DRIVER loading error. JDBC_DRIVER: " + JDBC_DRIVER);
      return null;
    } catch(SQLException se) {
      logger.fatal("SQL read error. JDBC_DRIVER: " + JDBC_DRIVER + ". MYSQL DB_URL: " + DB_URL);
      return null;
    } 

    return ImmutableListMultimap.<String, Ad>builder()
        .putAll("photography", aditemMap.get("camera"), aditemMap.get("lens"))
        .putAll("vintage", aditemMap.get("camera"), aditemMap.get("lens"), aditemMap.get("recordPlayer"))
        .put("cycling", aditemMap.get("bike"))
        .put("cookware", aditemMap.get("baristaKit"))
        .putAll("gardening", aditemMap.get("airPlant"), aditemMap.get("terrarium"))
        .build();
  }

  private class OpenTelemetryServerInterceptor implements io.grpc.ServerInterceptor {
    @Override
    public <ReqT, RespT> ServerCall.Listener<ReqT> interceptCall(
        ServerCall<ReqT, RespT> call, Metadata headers, ServerCallHandler<ReqT, RespT> next) {
      // Extract the Span Context from the metadata of the gRPC request
      Context extractedContext = textFormat.extract(Context.current(), headers, getter);
      InetSocketAddress clientInfo =
          (InetSocketAddress) call.getAttributes().get(Grpc.TRANSPORT_ATTR_REMOTE_ADDR);
      // Build a span based on the received context
      try (Scope scope = ContextUtils.withScopedContext(extractedContext)) {
        Span span =
            tracer
                .spanBuilder("hipstershop.AdService/GetAds")
                .setSpanKind(Span.Kind.SERVER)
                .startSpan();
        span.setAttribute("component", "grpc");
        span.setAttribute("rpc.service", "hipstershop.AdService");
        span.setAttribute("net.peer.ip", clientInfo.getHostString());
        span.setAttribute("net.peer.port", clientInfo.getPort());
        span.setAttribute("name", podName);
        span.setAttribute("node_name", nodeName);
        span.setAttribute("namespace", namespace);
        // Process the gRPC call normally
        try {
          span.setStatus(Status.OK);
          return Contexts.interceptCall(Context.current(), call, headers, next);
        } finally {
          span.end();
        }
      }
    }
  }

  /** Main launches the server from the command line. */
  public static void main(String[] args) throws IOException, InterruptedException {
    // Registers all RPC views.
    /*
     [TODO:rghetia] replace registerAllViews with registerAllGrpcViews. registerAllGrpcViews
     registers new views using new measures however current grpc version records against old
     measures. When new version of grpc (0.19) is release revert back to new. After reverting back
     to new the new measure will not provide any tags (like method). This will create some
     discrepencies when compared grpc measurements in Go services.
    */
    //RpcViews.registerAllViews();

    //new Thread(
    //        () -> {
    //          initStats();
    //          initTracing();
    //        })
    //    .start();
    // Start the RPC server. You shouldn't see any output from gRPC before this.
    logger.info("AdService starting.");
    final AdService service = AdService.getInstance();
    service.start();
    service.blockUntilShutdown();
  }
}
