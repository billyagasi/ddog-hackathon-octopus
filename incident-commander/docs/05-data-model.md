# docs/05-data-model.md

# PostgreSQL Schema

## incidents

* id
* incident_type
* severity
* status
* created_at
* closed_at

---

## findings

* id
* incident_id
* agent_name
* finding
* confidence
* created_at

---

## timeline

* id
* incident_id
* actor
* event
* timestamp

---

## recommendations

* id
* incident_id
* recommendation
* risk
* confidence

---

## approvals

* id
* incident_id
* approver
* status
* timestamp

---

## rca

* incident_id
* summary
* root_cause
* resolution
* lessons_learned

---

# Qdrant Collections

## incidents

Historical incidents.

## runbooks

Operational runbooks.

## lessons_learned

Knowledge retention.
