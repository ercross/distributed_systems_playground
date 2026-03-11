from dsl import Dashboard, Section, Target, ts, stat

SERVICES_RE = r'dsp-order-api|dsp-order-processor|dsp-payment-service'

DASH = Dashboard(
    title="Distributed Systems - Failure First (No Guards v0.1.0)",
    uid="dist-sys-failure-v010-tight",
    tags=["distributed-systems", "failure-first", "no-guards", "v0.1.0"],
    refresh="5s",
    time_from="now-30m",
    time_to="now",
    schema_version=39,
    version=1,
    sections=[
        Section(
            title="System Shape",
            teaching_note=(
                "This is the shortest truthful story of the system.\n"
                "Work enters. Edge latency rises. Forwarders pile up. Queue depth grows.\n"
                "If all four move together, the system is accepting work faster than it can turn into useful progress."
            ),
            panels=[
                ts(
                    "Incoming order requests at edge (req/s)",
                    targets=[
                        Target(
                            expr='sum(rate(order_api_http_requests_total{method="POST",path="/orders"}[30s]))',
                            legend="POST /orders req/s"
                        )
                    ],
                    w=6, h=6,
                    how_to_read="Incoming request rate at the edge for order creation.",
                    look_for="If this stays high while backlog and blocked work rise, the edge is still feeding a failing system.",
                ),
                ts(
                    "Edge request latency p95 (POST /orders)",
                    targets=[
                        Target(
                            expr='histogram_quantile(0.95, sum(rate(order_api_http_request_duration_seconds_bucket{method="POST",path="/orders"}[30s])) by (le))',
                            legend="p95"
                        )
                    ],
                    w=6, h=6,
                    how_to_read="User-visible tail latency at the edge.",
                    look_for="Rising p95 with continued intake is the first user-facing sign of internal waiting.",
                    common_traps="This depends on your RED middleware exposing order_api_http_request_duration_seconds_bucket.",
                ),
                ts(
                    "Forward goroutines in-flight",
                    targets=[
                        Target(
                            expr='order_api_processor_forward_in_flight',
                            legend="forward_in_flight"
                        )
                    ],
                    w=6, h=6,
                    how_to_read="How many edge goroutines are currently occupied trying to hand work to the processor.",
                    look_for="A steady rise means the edge is becoming a waiting room.",
                ),
                ts(
                    "Processor queue depth",
                    targets=[
                        Target(
                            expr='order_processor_queue_depth',
                            legend="queue_depth"
                        )
                    ],
                    w=6, h=6,
                    how_to_read="Current backlog waiting inside the processor.",
                    look_for="Depth trending upward is stored failure debt.",
                ),
            ],
        ),

        Section(
            title="Processor Accumulation",
            teaching_note=(
                "This section proves whether the processor is keeping up.\n"
                "The pattern to teach is simple: enqueue exceeds dequeue, queue age rises, and successful progress fades."
            ),
            panels=[
                ts(
                    "Enqueue vs dequeue rate",
                    targets=[
                        Target(
                            expr='rate(order_processor_orders_enqueued_total[30s])',
                            legend="enqueued/s"
                        ),
                        Target(
                            expr='rate(order_processor_orders_dequeued_total[30s])',
                            legend="dequeued/s"
                        ),
                    ],
                    w=8, h=7,
                    how_to_read="Compares admitted processor work to work actually pulled off the queue.",
                    look_for="If enqueued stays above dequeued, backlog grows.",
                ),
                ts(
                    "Processor in-flight work",
                    targets=[
                        Target(
                            expr='order_processor_in_flight',
                            legend="in_flight"
                        )
                    ],
                    w=8, h=7,
                    how_to_read="Orders actively being processed right now.",
                    look_for="If this rises but useful completions do not, concurrency is turning into overhead.",
                ),
                ts(
                    "Processor outcomes (rate)",
                    targets=[
                        Target(
                            expr='sum(rate(order_processor_orders_processed_total[30s])) by (outcome)',
                            legend='{{outcome}}'
                        )
                    ],
                    w=8, h=7,
                    how_to_read="Shows what completed work is becoming.",
                    look_for="Success flattening while enqueue remains high means incoming work is no longer turning into useful output.",
                ),
                ts(
                    "Seconds since last successful processing",
                    targets=[
                        Target(
                            expr='time() - order_processor_last_processed_timestamp_seconds',
                            legend="seconds_since_last_success"
                        )
                    ],
                    w=12, h=7,
                    how_to_read="How long it has been since the processor last made useful forward progress.",
                    look_for="If this rises continuously, the pipeline is alive but not useful.",
                ),
                ts(
                    "Payment call latency p95 from processor",
                    targets=[
                        Target(
                            expr='histogram_quantile(0.95, sum(rate(order_processor_payment_call_duration_seconds_bucket[30s])) by (le))',
                            legend="p95"
                        )
                    ],
                    w=12, h=7,
                    how_to_read="Tail latency of processor calls to payment.",
                    look_for="If this rises before processor success collapses, payment is likely the immediate bottleneck.",
                ),
                ts(
                    "Oldest queued work age (seconds)",
                    targets=[
                        Target(
                            expr='order_processor_oldest_queued_age_seconds',
                            legend="oldest_age_seconds"
                        )
                    ],
                    w=12, h=7,
                    how_to_read="Age of the oldest order still waiting in the queue.",
                    look_for="If this rises continuously, work is rotting in the queue even while the system remains technically alive.",
                ),
            ],
        ),

        Section(
            title="Dependency Blocking",
            teaching_note=(
                "This is where partial failure becomes slow failure.\n"
                "The most important question here is not just 'is payment failing?' but 'is payment trapping work in waiting states?'"
            ),
            panels=[
                stat(
                    "Payment service up",
                    targets=[
                        Target(expr='payment_service_up', legend="up")
                    ],
                    w=6, h=6,
                    thresholds_steps=[
                        {"color": "red", "value": 0},
                        {"color": "green", "value": 1},
                    ],
                    how_to_read="Basic liveness signal only.",
                    look_for="If this is 1 while progress stalls elsewhere, the service is alive but not useful.",
                ),
                stat(
                    "Bank hang enabled",
                    targets=[
                        Target(expr='payment_service_bank_hang_enabled', legend="hang_enabled")
                    ],
                    w=6, h=6,
                    thresholds_steps=[
                        {"color": "green", "value": 0},
                        {"color": "yellow", "value": 1},
                    ],
                    how_to_read="Experiment toggle for the injected slow or hanging failure mode.",
                    look_for="When this is 1, blocked concurrency should begin rising shortly after.",
                ),
                ts(
                    "Payment transactions in-flight",
                    targets=[
                        Target(
                            expr='payment_service_transactions_in_flight',
                            legend="transactions_in_flight"
                        )
                    ],
                    w=6, h=6,
                    how_to_read="Active requests inside payment.",
                    look_for="Rising in-flight means the service is holding more and more unfinished work.",
                ),
                ts(
                    "Bank hangs in-flight",
                    targets=[
                        Target(
                            expr='payment_service_bank_hangs_in_flight',
                            legend="hangs_in_flight"
                        )
                    ],
                    w=6, h=6,
                    how_to_read="Requests currently stuck in the simulated hanging dependency.",
                    look_for="A rising line is work entering time without exiting it.",
                ),
                ts(
                    "Payment outcomes (increase over 30w)",
                    targets=[
                        Target(
                            expr='sum(increase(payment_service_outcomes_total[30s])) by (outcome)',
                            legend='{{outcome}}'
                        )
                    ],
                    w=12, h=7,
                    how_to_read="High-level completion mix inside payment over the last 1 minutes.",
                    look_for="Success flattening while in-flight and hangs rise indicates trapped work rather than clean failure.",
                ),
                ts(
                    "Payment processing latency p95",
                    targets=[
                        Target(
                            expr='histogram_quantile(0.95, sum(rate(payment_service_processing_duration_seconds_bucket[30s])) by (le))',
                            legend="p95"
                        )
                    ],
                    w=12, h=7,
                    how_to_read="Tail latency for payment processing.",
                    look_for="Useful while requests still complete. Less truthful once hang-dominated behavior takes over.",
                ),
                ts(
                    "Bank gateway latency p95",
                    targets=[
                        Target(
                            expr='histogram_quantile(0.95, sum(rate(payment_service_bank_gateway_duration_seconds_bucket[30s])) by (le))',
                            legend="p95"
                        )
                    ],
                    w=12, h=7,
                    how_to_read="Tail latency specifically around the simulated bank dependency.",
                    look_for="Confirms whether the dependency itself is where time is being lost.",
                ),
            ],
        ),

        Section(
            title="Runtime Consequences",
            teaching_note=(
                "These are consequence panels, not root-cause panels.\n"
                "Read them after you already established the blocking chain.\n"
                "For this lesson, goroutines matter more than deep GC internals."
            ),
            panels=[
                ts(
                    "Goroutines by service",
                    targets=[
                        Target(
                            expr=f'max by (job) (go_goroutines{{job=~"{SERVICES_RE}"}})',
                            legend='{{job}}'
                        )
                    ],
                    w=12, h=7,
                    how_to_read="Direct view of goroutine accumulation across services.",
                    look_for="Rising goroutines alongside rising in-flight and queue depth is the runtime fingerprint of quiet drowning.",
                ),
                ts(
                    "RSS memory by service",
                    targets=[
                        Target(
                            expr=f'max by (job) (process_resident_memory_bytes{{job=~"{SERVICES_RE}"}})',
                            legend='{{job}}'
                        )
                    ],
                    w=12, h=7,
                    how_to_read="Resident memory as seen by the OS.",
                    look_for="If this climbs steadily during accumulation, waiting work is becoming real process pressure.",
                ),
                ts(
                    "CPU usage (cores) by service",
                    targets=[
                        Target(
                            expr=f'sum by (job) (rate(process_cpu_seconds_total{{job=~"{SERVICES_RE}"}}[30s]))',
                            legend='{{job}}'
                        )
                    ],
                    w=12, h=7,
                    how_to_read="CPU consumption by each process.",
                    look_for="If latency and backlog rise without matching CPU growth, the system is likely waiting on dependency latency rather than burning compute.",
                ),
            ],
        ),
    ],
)

READING_GUIDE = """
How to read this dashboard

Narrative arc:
1) Work keeps entering at the edge.
2) Forwarders begin piling up.
3) Queue depth grows because enqueue stays above dequeue.
4) Useful progress slows or stops.
5) Payment traps more work in-flight and in hangs.
6) Goroutines and memory begin to rise as runtime consequences.

Operational reading checklist:
A) Is work still entering? (Incoming order requests at edge)
B) Are users feeling slowdown? (Edge p95)
C) Is the edge blocking on downstream? (Forward goroutines in-flight)
D) Is backlog growing? (Enqueue vs dequeue, queue depth, oldest queued age)
E) Is the processor still making useful progress? (Outcomes, seconds since last success)
F) Is payment trapping work? (Transactions in-flight, bank hangs in-flight)
G) Is the dependency the immediate bottleneck? (Processor payment p95, bank gateway p95)
H) What is the runtime cost? (Goroutines, RSS, CPU)

What this dashboard intentionally does not optimize for:
- Full HTTP breakdown everywhere
- Rich error taxonomies
- Deep GC analysis
- Broad service health theater

This is a failure-first dashboard.
Its job is to make one sentence visually undeniable:
the system keeps accepting work while unfinished work accumulates and useful progress fades.
"""