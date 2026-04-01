### Fleet Evaluator

This is a work in progress, I use Claude and Gemini to produce some skills based on the Fleet documentation, then write/produce some code that evaluates fleet resources against best practices, with an option to keep the report in a nats kv or jetstream store.  This way, endusers can subscribe to the reports they need, and multi-tenancy can be handled simply with nats user access.

---
