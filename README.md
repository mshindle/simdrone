# simdrone: The Drone Army Simulator 🚁

Let's dive into **simdrone**, a fascinating piece of software designed to simulate a fleet of drones and how they communicate. Think of it as a laboratory where we study how complex systems talk to each other, process information, and store data in real-time.

---

## 🎓 What is Simdrone?

Simdrone is a distributed system that mimics a drone army. It isn't just about one single program; it's a collection of specialized services that work together. Its main goal is to demonstrate modern software patterns like **Event-Driven Architecture**, **Dependency Injection**, and **Microservices**.

### The Core Concept

Imagine hundreds of drones flying in the sky. Each drone needs to report its position, its battery level, and any faults it encounters. Simdrone provides the infrastructure to receive these reports, "understand" them, and save them for later analysis.

---

## 🏗️ How it Works (The Architecture)

Our "drone army" is built with three main pillars:

### 1. The Gateway (The `handler` Service)
The `handler` is our front door. It exposes an HTTP API (using the **Echo** framework) that drones "call" to report their status. 
- It validates the incoming data (making sure latitude isn't 500, for example!).
- It then converts these HTTP requests into **Events**.
- Finally, it hands off these events to the "Nervous System."

### 2. The Nervous System (NATS & JetStream)
Communication is key! Instead of services talking directly to each other (which is fragile), they use a **Message Bus**. 
- We primarily use **NATS JetStream** as our high-performance message broker.
- When the `handler` gets a status report, it "dispatches" an event onto a specific topic (like `events.drone.telemetry.updated`).
- This decouples our services: the `handler` doesn't care who is listening; it just knows it sent the message.

### 3. The Brain (The `process` Service)
The `process` service is our event consumer. It listens to the "Nervous System" for any new drone events.
- When it hears an event, it processes it (like rolling up statistics).
- It then stores the data into **MongoDB**, our persistent memory.

### 4. The Eyes (The `view` Service)
The `view` service is our query engine. It provides an HTTP API to retrieve the latest state of any drone.
- It queries **MongoDB** to find the most recent telemetry, position, or alert for a specific drone ID.
- This allows operators to see the "Current State" of the fleet.

---

## 🏛️ A Note on CQRS
While this application demonstrates **CQRS (Command Query Responsibility Segregation)** by separating our "Write" path (`handler` and `process`) from our "Read" path (`view`), it currently contains a classic architectural shortcut. 

In a strict CQRS implementation, the "Read" side should ideally have its own data store (an "Eventually Consistent" view) optimized for queries. Currently, both the `process` and `view` services share the same MongoDB collections. This means our database is the point of integration—a violation we've accepted for this demo, but something to keep in mind for production systems!

---

## 🚀 Getting Started

Ready to fly? Let's get the environment up and running.

### 📋 Prerequisites
You'll need **Docker** and **Go 1.26** installed on your machine.

### 1. Start the Infrastructure
We use `docker-compose` to spin up our dependencies (NATS and MongoDB) quickly.

```bash
docker-compose up -d
```

### 2. Run the Handler
In a new terminal, start the gateway service:

```bash
go run main.go handler
```
*The handler starts on port 8080 by default.*

### 3. Run the Processor
In another terminal, start the brain:

```bash
go run main.go process
```

### 4. Run the Viewer
Finally, in one more terminal, start the eyes:

```bash
go run main.go view
```
*The viewer starts on port 1315 by default.*

---

## 🧪 Testing and Verification

To ensure our drone army is behaving correctly, we have several ways to test:

### Unit Tests
We have comprehensive unit tests for our logic. You can run them all with:

```bash
go test ./...
```

### Manual Verification (The "Drone" Simulation)
You can simulate a drone sending a command using `curl`. For example, to report telemetry:

```bash
curl -X POST http://localhost:8080/api/cmds/telemetry \
  -H "Content-Type: application/json" \
  -d '{"drone_id":"drone-01", "battery": 85, "uptime": 1200, "core_temp": 35}'
```

If everything is working, the `handler` will return `201 Created`, the `process` service will log the event, and it will be saved to MongoDB!

### Querying the State (Using the Viewer)
Once you've sent some data, you can use the `view` service to see the latest reports:

```bash
# Get the last telemetry for drone-01
curl http://localhost:1315/drones/drone-01/lastTelemetry

# Get the last position for drone-01
curl http://localhost:1315/drones/drone-01/lastPosition

# Get the last alert for drone-01
curl http://localhost:1315/drones/drone-01/lastAlert
```

---
*Happy Simulating! If you have questions, remember: in an event-driven world, everything is just a message away.*

