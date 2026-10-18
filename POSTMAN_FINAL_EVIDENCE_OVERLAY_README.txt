OrderFlow final Postman/evidence overlay

Extract at the repository root. Run these scripts from the repository root:
  python3 scripts/finalize-postman-collection-name.py
  python3 scripts/register-postman-evidence.py

The new neutral collection file becomes:
  docs/postman/OrderFlow_API_Verification.postman_collection.json

Only import that JSON collection in Postman after running the collection rename script. It changes the displayed name to "OrderFlow API Verification — Local"; it does not change endpoints, payloads, tests or variables.

Included screenshots are:
- successful public product list 200
- Postman collection-folder layout (reference only, not final API evidence)
- historical successful CI run for SHA 2654070 (recapture CI after final evidence commit)

The file docs/evidence/POSTMAN_FINAL_CAPTURE_PLAN.md explains the UPDATE 0 admin fix, order creation/order-history steps, smallest recommended screenshot set, and Swagger/OpenAPI status.
Do not commit raw JWT screenshots, .env, credentials, database dumps or populated Postman environment exports.
