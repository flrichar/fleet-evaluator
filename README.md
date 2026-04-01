### Fleet Evaluator

This is a work in progress, I used Claude and Gemini to produce some skills based on the Fleet documentation, then write some code that can evaluate fleet resources against best practices, with an option to keep the report in a nats kv or jetstream store.  This way, endusers could subsribe to the reports they need, and multi-tenancy can be handled simply with nats user access.

---
