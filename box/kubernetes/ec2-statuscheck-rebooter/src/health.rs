use anyhow::Result;
use http_body_util::Full;
use hyper::body::Bytes;
use hyper::server::conn::http1;
use hyper::service::service_fn;
use hyper::{Request, Response, StatusCode};
use hyper_util::rt::TokioIo;
use std::net::SocketAddr;
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};
use tokio::net::TcpListener;
use tracing::{debug, error, info};

/// Health check server state
#[derive(Clone)]
pub struct HealthServer {
    ready: Arc<AtomicBool>,
}

impl Default for HealthServer {
    fn default() -> Self {
        Self::new()
    }
}

impl HealthServer {
    pub fn new() -> Self {
        Self {
            ready: Arc::new(AtomicBool::new(false)),
        }
    }

    /// Mark the application as ready
    pub fn set_ready(&self, ready: bool) {
        self.ready.store(ready, Ordering::SeqCst);
        if ready {
            info!("Application marked as ready");
        }
    }

    /// Start the HTTP health check server
    pub async fn serve(
        self,
        port: u16,
        ready_signal: tokio::sync::oneshot::Sender<()>,
    ) -> Result<()> {
        let addr = SocketAddr::from(([0, 0, 0, 0], port));
        let listener = TcpListener::bind(addr).await?;

        info!(
            port = port,
            address = %addr,
            liveness_endpoint = format!("http://0.0.0.0:{}/healthz", port),
            readiness_endpoint = format!("http://0.0.0.0:{}/readyz", port),
            "Health check server started successfully"
        );

        // Signal that the server is ready to accept connections
        let _ = ready_signal.send(());

        loop {
            let (stream, peer_addr) = match listener.accept().await {
                Ok(conn) => conn,
                Err(e) => {
                    error!(error = %e, "Failed to accept connection");
                    continue;
                }
            };

            let io = TokioIo::new(stream);
            let server = self.clone();

            tokio::spawn(async move {
                if let Err(e) = http1::Builder::new()
                    .serve_connection(
                        io,
                        service_fn(move |req| {
                            let server = server.clone();
                            async move { server.handle_request(req).await }
                        }),
                    )
                    .await
                {
                    debug!(
                        peer = %peer_addr,
                        error = %e,
                        "Error serving connection"
                    );
                }
            });
        }
    }

    async fn handle_request(
        &self,
        req: Request<hyper::body::Incoming>,
    ) -> Result<Response<Full<Bytes>>, hyper::Error> {
        let path = req.uri().path();

        debug!(
            method = %req.method(),
            path = %path,
            "Received health check request"
        );

        match path {
            "/healthz" | "/health" => Ok(self.handle_liveness()),
            "/readyz" | "/ready" => Ok(self.handle_readiness()),
            _ => Ok(Response::builder()
                .status(StatusCode::NOT_FOUND)
                .body(Full::new(Bytes::from("Not Found")))
                .unwrap()),
        }
    }

    fn handle_liveness(&self) -> Response<Full<Bytes>> {
        // Always return 200 if the process is running
        Response::builder()
            .status(StatusCode::OK)
            .body(Full::new(Bytes::from("ok")))
            .unwrap()
    }

    fn handle_readiness(&self) -> Response<Full<Bytes>> {
        if self.ready.load(Ordering::SeqCst) {
            Response::builder()
                .status(StatusCode::OK)
                .body(Full::new(Bytes::from("ready")))
                .unwrap()
        } else {
            Response::builder()
                .status(StatusCode::SERVICE_UNAVAILABLE)
                .body(Full::new(Bytes::from("not ready")))
                .unwrap()
        }
    }
}
