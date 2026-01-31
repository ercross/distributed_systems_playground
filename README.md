# Distributed Systems Playground - Failure First

A hands-on playground for learning distributed systems by intentionally building fragile systems and evolving them through
constraint-driven design.

It is intentionally small, local, and understandable end-to-end.
There is no Kubernetes, no cloud magic, and no hidden abstractions.

The goal is not to “build a system”.
The goal is to experience how systems fail, and then earn reliability one constraint at a time.

Although the example domain is an event-driven order flow,
the lessons here apply to all distributed systems:
RPC systems, microservices, databases, schedulers, and control planes.


## Core Philosophy

> **Design the system to fail first.  
> Then introduce constraints deliberately and observe how behavior changes.**

Most systems are only tested in their “happy path”.
This project starts at the opposite end.

You will feel:
- partial failure
- tail latency explosions
- retry amplification
- queue pile-ups
- cascading failures
- recovery through backpressure and idempotency

This is not theory.  
Every lesson is backed by load tests and observable behavior.


## Versioning

This repository uses semantic versioning to represent system evolution.

Each version introduces exactly one new constraint and captures the system’s behavior at that stage.

Versions are not feature releases.
They are design states.

This mirrors how real systems evolve in production.


## Versioning Model

### v0.x — Failure-First Era

Versions in `v0.x` are **intentionally unstable by design**.

They exist to demonstrate:
- what breaks
- why it breaks
- how naïve designs collapse under load

### v1.0.0 — Earned Stability

`v1.0.0` will only exist once:
- failures are bounded
- backpressure is explicit
- retries are controlled
- latency remains stable under overload


## Planned Version Progression

| Version | Constraint Introduced | Purpose |
|------|-----------------------|---------|
| v0.1.0 | Naive baseline | Establish failure |
| v0.2.0 | Admission control | Stop accepting infinite work |
| v0.3.0 | Backpressure | Bound concurrency |
| v0.4.0 | Timeouts | Prevent resource leaks |
| v0.5.0 | Idempotency | Make retries safe |
| v0.6.0 | Bounded retries | Prevent retry storms |
| v0.7.0 | Circuit breaking | Contain blast radius |
| v0.8.0 | Observability | See failure early |
| v1.0.0 | Composed system | Survives partial failure |

Each version:
- changes **as little code as possible**
- keeps business logic the same
- reruns the **same load tests**
- compares behavior against previous versions


## How Versioning Works in This Repo

### Git Tags as Snapshots

Each version is captured using a **Git tag**.

Example:
```bash
git checkout v0.1.0
````

Tags are:

* immutable
* reproducible
* exact snapshots of system behavior

This allows you to:

* reproduce failures exactly
* compare versions under identical load
* understand cause → effect clearly

There are **no long-lived branches** per version.
The `main` branch always represents the latest evolution.


## What This Project Intentionally Avoids

To keep the learning signal strong, this project avoids:

* Kubernetes
* autoscaling
* managed queues
* cloud services
* heavy frameworks

If you don’t understand *why* something happens here, adding more infrastructure will only hide it.


## Who This Is For

This project is ideal for:

* backend engineers moving into distributed systems
* senior engineers who want sharper failure intuition
* anyone preparing for system design interviews
* engineers tired of “happy path” demos


## Final Note

Most distributed systems don’t fail loudly.

They fail:

* slowly
* partially
* convincingly

This project exists to make that failure impossible to ignore — and then to show how reliability is earned, not added.


Follow the versions.
Feel the failures.
Earn the fixes.
