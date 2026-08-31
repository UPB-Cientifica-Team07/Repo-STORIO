class MetricsState {
  constructor() {
    this.activeConnections = 0;
    this.totalRequests = 0;
  }

  connectionOpened() {
    this.activeConnections++;
  }

  connectionClosed() {
    if (this.activeConnections > 0) {
      this.activeConnections--;
    }
  }

  requestReceived() {
    this.totalRequests++;
  }

  getActiveConnections() {
    return this.activeConnections;
  }

  getTotalRequests() {
    return this.totalRequests;
  }
}

module.exports = new MetricsState();
