ORDERFLOW POSTMAN AUTH-FIX COLLECTION

Import only this JSON file into Postman:
  docs/postman/OrderFlow_Guided_Evidence.postman_collection.json

Do not import a second collection from another pack. If the older guided collection is already imported, delete that older Postman collection first, then import the auth-fix JSON from this ZIP.

After importing, run Folder 01 through `05 — Login existing ADMIN`. Check the response role is ADMIN. Then run Folder 02; `GET /products` is public (200), but create/update/deactivate require ADMIN.

Never commit tokens/secrets or screenshots with tokens visible.
