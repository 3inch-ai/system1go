# Live verification boundary

No live local model, Jev or OpenAI inference was performed for system1go. All HTTP exercises use synthetic loopback httptest handlers. No credentials were read, paid inference requests sent, servers started/stopped or model weights downloaded. The parent TypeScript repository has a historical H2O smoke record, which does not validate this Go port.

A future authorized smoke should record exact SDK SHA, endpoint protocol, server/shim SHA, model/quantization identity, one bounded synthetic batched request, original versus normalized response checks and remaining limitations. Never make it a default test, import side effect or CI step. Do not discover credentials or operate another session's server. Provider acceptance, calibration, accuracy and performance require separate evidence.
