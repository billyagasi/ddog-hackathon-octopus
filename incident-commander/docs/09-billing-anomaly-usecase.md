# docs/09-billing-anomaly-usecase.md

# Use Case

AWS Billing Anomaly

## Scenario

Expected Cost:
$300/day

Current Cost:
$1,200/day

## Investigation

Infrastructure AI:

* Aurora scale-up

Application AI:

* retry storm

Incident Management AI:

* projected monthly impact
  $27,000

## Recommendation

Reduce Aurora size.

Review retry policy.

Review HPA settings.
