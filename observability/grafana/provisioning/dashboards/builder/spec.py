from dsl import Dashboard, Section, Target, ts, stat

SERVICES_RE = r'dsp-order-api|dsp-order-processor|dsp-payment-service'

DASH = Dashboard(
    title="Distributed Systems - Failure First (v0.2.0 Admission Control)",
    uid="dist-sys-failure-v020-admission-control",
    tags=["distributed-systems", "failure-first", "admission-control", "v0.2.0"],
    refresh="5s",
    time_from="now-30m",
    time_to="now",
    schema_version=39,
    version=1,
    sections=[
        Section(
            title="Ingress Admission Control",
            teaching_note=(
                "v0.2.0 changes the edge contract.\n"
                "The system should no longer accept unlimited work.\n"
                "The key questions here are: are admission slots saturating, and does the edge begin rejecting excess work on purpose?"
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
                    w=8, h=6,
                    how_to_read="Incoming request rate at the edge for order creation.",
                    look_for="If this stays high while admission usage reaches the limit, the edge is under sustained pressure.",
                ),
                ts(
                    "Admission slots in use vs configured limit",
                    targets=[
                        Target(
                            expr='order_api_admission_in_use',
                            legend="in_use"
                        ),
                        Target(
                            expr='order_api_admission_limit',
                            legend="limit"
                        ),
                    ],
                    w=8, h=6,
                    how_to_read="Authoritative admission occupancy versus configured capacity.",
                    look_for="When in_use reaches and sticks near limit, new work should start getting rejected instead of being admitted forever.",
                ),
                ts(
                    "Accepted vs rejected orders (rate)",
                    targets=[
                        Target(
                            expr='rate(order_api_orders_accepted_total[30s])',
                            legend="accepted/s"
                        ),
                        Target(
                            expr='rate(order_api_orders_rejected_total[30s])',
                            legend="rejected/s"
                        ),
                    ],
                    w=8, h=6,
                    how_to_read="High-level edge decision split.",
                    look_for="Under overload, accepted rate should flatten while rejected rate rises. That is controlled degradation, not unlimited intake.",
                ),
                ts(
                    "Admission vs validation rejections (rate)",
                    targets=[
                        Target(
                            expr='rate(order_api_admission_rejections_total[30s])',
                            legend="admission_rejections/s"
                        ),
                        Target(
                            expr='rate(order_api_validation_rejections_total[30s])',
                            legend="validation_rejections/s"
                        ),
                    ],
                    w=12, h=7,
                    how_to_read="Separates intentional overload shedding from ordinary bad input.",
                    look_for="The lesson signal is admission rejections rising because slots are exhausted, not validation failures rising.",
                ),
                ts(
                    "Forwarding goroutines in-flight",
                    targets=[
                        Target(
                            expr='order_api_processor_forward_in_flight',
                            legend="forward_in_flight"
                        )
                    ],
                    w=6, h=7,
                    how_to_read="Accepted orders whose forwarding lifecycle is still active.",
                    look_for="If this stays elevated, admission slots are being held by unfinished forwarding work.",
                ),
                ts(
                    "Forwarding duration p95",
                    targets=[
                        Target(
                            expr='histogram_quantile(0.95, sum(rate(order_api_processor_forward_duration_seconds_bucket[30s])) by (le))',
                            legend="p95"
                        )
                    ],
                    w=6, h=7,
                    how_to_read="Tail time spent trying to hand accepted work to the processor.",
                    look_for="Rising p95 means admitted work is taking longer to escape the edge, so slots free up more slowly.",
                    common_traps="This is forwarding lifecycle time, not full business completion time.",
                ),
            ],
        ),

        Section(
            title="Internal Flow Without Back-Pressure",
            teaching_note=(
                "Admission control is present, but internal pacing is still missing.\n"
                "This section should reveal the core v0.2.0 tension:\n"
                "the system stops accepting infinite work at ingress, yet admitted work can still fan out uncontrollably inside."
            ),
            panels=[
                ts(
                    "Processor enqueue vs dequeue rate",
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
                    how_to_read="Compares admitted processor work to work removed from the queue.",
                    look_for="If enqueue remains above dequeue, backlog grows. If dequeue stays high but in-flight explodes, work is being drained into unbounded execution.",
                ),
                ts(
                    "Processing goroutines spawned (rate)",
                    targets=[
                        Target(
                            expr='rate(order_processor_processing_spawned_total[30s])',
                            legend="spawned/s"
                        )
                    ],
                    w=8, h=7,
                    how_to_read="How fast the processor turns dequeued work into goroutines.",
                    look_for="A high spawn rate with no worker bound is the internal fan-out signature of missing back-pressure.",
                ),
                ts(
                    "Processor in-flight vs payment calls in-flight",
                    targets=[
                        Target(
                            expr='order_processor_in_flight',
                            legend="processor_in_flight"
                        ),
                        Target(
                            expr='order_processor_payment_calls_in_flight',
                            legend="payment_calls_in_flight"
                        ),
                    ],
                    w=8, h=7,
                    how_to_read="Where admitted work is currently sitting inside the processor.",
                    look_for="If both rise together, the processor is not pacing work before hitting payment.",
                ),
                ts(
                    "Processor queue depth",
                    targets=[
                        Target(
                            expr='order_processor_queue_depth',
                            legend="queue_depth"
                        )
                    ],
                    w=8, h=7,
                    how_to_read="Current number of orders waiting in the queue.",
                    look_for="Useful, but not sufficient. In v0.2.0 this may stay deceptively low because work is drained immediately into goroutines.",
                    common_traps="Do not conclude the system is healthy just because queue depth is low.",
                ),
                ts(
                    "Oldest queued work age (seconds)",
                    targets=[
                        Target(
                            expr='order_processor_oldest_queued_age_seconds',
                            legend="oldest_age_seconds"
                        )
                    ],
                    w=8, h=7,
                    how_to_read="Age of the oldest order still waiting in the queue.",
                    look_for="If this rises, queueing delay is visible. If it stays low while in-flight explodes, the backlog is hiding in execution instead of the queue.",
                ),
                ts(
                    "Queue wait duration p95",
                    targets=[
                        Target(
                            expr='histogram_quantile(0.95, sum(rate(order_processor_queue_wait_duration_seconds_bucket[30s])) by (le))',
                            legend="p95"
                        )
                    ],
                    w=8, h=7,
                    how_to_read="Tail wait time before processing begins.",
                    look_for="Rising queue wait confirms stored backlog. A flat line here does not disprove overload if execution in-flight is climbing.",
                ),
                ts(
                    "Processor outcomes (rate)",
                    targets=[
                        Target(
                            expr='sum(rate(order_processor_orders_processed_total[30s])) by (outcome)',
                            legend='{{outcome}}'
                        )
                    ],
                    w=12, h=7,
                    how_to_read="What completed processor work is becoming.",
                    look_for="If success flattens while spawn rate and in-flight rise, concurrency is turning into waiting rather than useful progress.",
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
                    how_to_read="How long it has been since the processor last produced a successful completion.",
                    look_for="A steadily rising line means the pipeline is still active but no longer making useful forward progress.",
                ),
            ],
        ),

        Section(
            title="Downstream Pressure Chain",
            teaching_note=(
                "This section connects the saturation story across services.\n"
                "The point is not merely that payment is slow.\n"
                "The point is that slow or hanging downstream work traps concurrency upstream and keeps admission slots occupied."
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
                    look_for="If this is 1 while progress elsewhere stalls, the service is alive but not necessarily useful.",
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
                    how_to_read="Experiment toggle for the injected hanging dependency behavior.",
                    look_for="When this is enabled, blocked concurrency should begin rising across the chain shortly after.",
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
                    look_for="A rising line means payment is holding unfinished work longer.",
                ),
                ts(
                    "Bank gateway calls in-flight",
                    targets=[
                        Target(
                            expr='payment_service_bank_gateway_calls_in_flight',
                            legend="gateway_calls_in_flight"
                        )
                    ],
                    w=6, h=6,
                    how_to_read="Requests currently inside the simulated bank gateway region.",
                    look_for="If this rises with processor payment calls in-flight, the immediate bottleneck is inside the dependency path.",
                ),
                ts(
                    "Bank hangs in-flight",
                    targets=[
                        Target(
                            expr='payment_service_bank_hangs_in_flight',
                            legend="hangs_in_flight"
                        )
                    ],
                    w=12, h=7,
                    how_to_read="Requests currently stuck in the simulated unbounded bank hang.",
                    look_for="A rising line is work entering time without exiting it. That is exactly the kind of trapped concurrency that should propagate upstream.",
                ),
                ts(
                    "Processor payment latency p95 vs bank gateway latency p95",
                    targets=[
                        Target(
                            expr='histogram_quantile(0.95, sum(rate(order_processor_payment_call_duration_seconds_bucket[30s])) by (le))',
                            legend="processor_to_payment_p95"
                        ),
                        Target(
                            expr='histogram_quantile(0.95, sum(rate(payment_service_bank_gateway_duration_seconds_bucket[30s])) by (le))',
                            legend="bank_gateway_p95"
                        ),
                    ],
                    w=12, h=7,
                    how_to_read="Correlates upstream dependency call latency with downstream bank-region latency.",
                    look_for="If both rise together, the processor is not the primary source of delay; it is inheriting delay from payment and the bank simulation.",
                ),
                ts(
                    "Payment outcomes (increase over 30s)",
                    targets=[
                        Target(
                            expr='sum(increase(payment_service_outcomes_total[30s])) by (outcome)',
                            legend='{{outcome}}'
                        )
                    ],
                    w=12, h=7,
                    how_to_read="High-level completion mix inside payment over the last 30 seconds.",
                    look_for="Success flattening while in-flight and hangs rise indicates trapped work rather than clean, fast failure.",
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
                    how_to_read="Tail latency for end-to-end payment handler time among requests that still complete.",
                    look_for="Useful while requests still exit. Less truthful once hang-dominated behavior takes over.",
                ),
            ],
        ),

        Section(
            title="Runtime Consequences",
            teaching_note=(
                "These are consequence panels, not control panels.\n"
                "By this point you should already know whether admission is working and where work is getting trapped.\n"
                "These panels show the runtime cost of that trapped work."
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
                    look_for="Rising goroutines, especially with stable or falling useful completions, is the runtime fingerprint of missing back-pressure.",
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
                    look_for="If this climbs while work is trapped in-flight, waiting concurrency is becoming real process pressure.",
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
                    look_for="If latency and in-flight rise without matching CPU growth, the system is likely waiting on dependency latency rather than burning compute.",
                ),
            ],
        ),
    ],
)

READING_GUIDE = """
How to read this dashboard

Narrative arc:
1) The edge receives sustained traffic.
2) Admission slots fill up to a fixed limit.
3) Excess work is rejected intentionally at ingress.
4) Already-admitted work still fans out internally without pacing.
5) Processor and payment in-flight work rise together.
6) Useful completions flatten or stop.
7) Goroutines and memory become the runtime cost of trapped work.

Operational reading checklist:
A) Is traffic arriving? (Incoming order requests at edge)
B) Is admission control visibly enforcing a boundary? (Admission slots in use vs configured limit)
C) Are rejections overload-driven rather than validation-driven? (Admission vs validation rejections)
D) Is accepted work getting stuck at the edge? (Forwarding goroutines in-flight, Forwarding duration p95)
E) Is internal work being paced, or just fanned out? (Processing goroutines spawned, Processor in-flight vs payment calls in-flight)
F) Where is backlog living? (Queue depth, oldest queued age, queue wait p95)
G) Is useful progress fading? (Processor outcomes, seconds since last successful processing)
H) Is downstream pressure the immediate cause? (Payment transactions in-flight, bank gateway calls in-flight, bank hangs in-flight, latency p95s)
I) What is the runtime cost? (Goroutines, RSS, CPU)

What this dashboard intentionally does not optimize for:
- Rich error taxonomies beyond what the lesson needs
- Full HTTP breakdown for every service
- Deep GC or allocator analysis
- Generic health theater

This is a teaching dashboard for v0.2.0.

Its job is to make two sentences visually undeniable:
1) admission control is present at ingress
2) missing back-pressure still lets admitted work accumulate and stall inside the system
"""