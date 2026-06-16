# Datadog MCP Strategy

## Core Principle

Every AI investigation must use Datadog MCP.

Direct infrastructure access is prohibited.

---

# Incident Tools

- get_incidents()
- get_incident()
- get_incident_timeline()
- get_incident_severity()

---

# Watchdog Tools

- get_watchdog_alerts()
- get_watchdog_rca()
- get_watchdog_impact()

---

# Service Catalog Tools

- get_service()
- get_dependencies()
- get_owner()
- get_scorecard()

---

# Observability Tools

- query_logs()
- query_traces()
- query_apm()
- query_error_tracking()

---

# Infrastructure Tools

- query_hosts()
- query_containers()
- query_kubernetes()
- query_database_monitoring()

---

# Change Intelligence

- query_deployments()
- query_change_tracking()
- query_events()

---

# Reliability Tools

- query_slo()
- query_error_budget()
- query_burn_rate()

---

# Cost Tools

- query_cloud_cost()
- query_cost_anomalies()

---

# Workflow Automation

- execute_workflow()
- rollback_deployment()
- restart_service()
- scale_service()