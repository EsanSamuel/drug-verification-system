# Drug Verification and Management System — API Documentation

## Overview
The **Computerized Drug Verification and Management System** API provides authentication, drug registration, manufacturer tracking, unique serialized drug unit tracking, public verification via serial number and QR code scanning, and verification audit trails.

**Base URL**: `/api/v1`

---

## Authentication & Authorization
All endpoints except `/health`, `/auth/register`, `/auth/login`, and `/verify/:serial` require a Bearer token in the `Authorization` header:

```http
Authorization: Bearer <jwt_token>
```

### Roles
- `pharmacist`: Registered pharmacist user authorized to manage drugs, batches, units, and inspect audit logs.
- `admin`: Administrative authority.

---

## Standard Response Format

### Success Response
```json
{
  "success": true,
  "data": { ... }
}
```

### Paginated Response
```json
{
  "success": true,
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": "Descriptive error message"
}
```

---

## Endpoints

### 1. Health & Status

#### `GET /api/v1/health`
Checks server health status.
* **Auth**: Public
* **Response `200 OK`**:
```json
{
  "status": "ok"
}
```

---

### 2. Authentication

#### `POST /api/v1/auth/register`
Registers a new pharmacist account and returns an auth token.
* **Auth**: Public
* **Request Body**:
```json
{
  "full_name": "Dr. Sarah Connor",
  "email": "sarah.connor@hospital.org",
  "password": "StrongPassword123!"
}
```
* **Response `201 Created`**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsIn...",
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "full_name": "Dr. Sarah Connor",
      "email": "sarah.connor@hospital.org",
      "role": "pharmacist"
    }
  }
}
```

#### `POST /api/v1/auth/login`
Authenticates existing pharmacist credentials.
* **Auth**: Public
* **Request Body**:
```json
{
  "email": "sarah.connor@hospital.org",
  "password": "StrongPassword123!"
}
```
* **Response `200 OK`**: Same structure as registration response.

#### `GET /api/v1/auth/me`
Retrieves details of the currently authenticated pharmacist.
* **Auth**: Bearer Token
* **Response `200 OK`**:
```json
{
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "full_name": "Dr. Sarah Connor",
    "email": "sarah.connor@hospital.org",
    "role": "pharmacist"
  }
}
```

---

### 3. Manufacturers

#### `POST /api/v1/manufacturers`
Registers a pharmaceutical manufacturer.
* **Auth**: Bearer Token
* **Request Body**:
```json
{
  "name": "GlaxoSmithKline Pharmaceuticals",
  "address": "12 Industrial Layout, Lagos, Nigeria"
}
```
* **Response `201 Created`**:
```json
{
  "success": true,
  "data": {
    "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "name": "GlaxoSmithKline Pharmaceuticals",
    "address": "12 Industrial Layout, Lagos, Nigeria",
    "created_at": "2026-09-23T07:00:00Z",
    "updated_at": "2026-09-23T07:00:00Z"
  }
}
```

#### `GET /api/v1/manufacturers`
Lists manufacturers with pagination.
* **Auth**: Bearer Token
* **Query Params**: `page` (default: 1), `limit` (default: 20)

#### `GET /api/v1/manufacturers/:id`
Retrieves details of a manufacturer.
* **Auth**: Bearer Token

#### `PATCH /api/v1/manufacturers/:id`
Updates manufacturer details.
* **Auth**: Bearer Token

#### `DELETE /api/v1/manufacturers/:id`
Deletes a manufacturer (fails if drugs are linked).
* **Auth**: Bearer Token

---

### 4. Drugs & Batches

#### `POST /api/v1/drugs`
Registers a new drug batch.
* **Auth**: Bearer Token
* **Request Body**:
```json
{
  "name": "Augmentin 625mg Tablets",
  "generic_name": "Amoxicillin / Clavulanic Acid",
  "manufacturer_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "batch_number": "AUG-2026-09",
  "nafdac_number": "A4-0123",
  "manufacturing_date": "2026-01-15",
  "expiry_date": "2028-01-15",
  "quantity": 5000,
  "status": "active"
}
```
* **Response `201 Created`**: Returns created drug object.

#### `GET /api/v1/drugs`
Search and list drugs with filtering.
* **Auth**: Bearer Token
* **Query Params**:
  - `page`, `limit`
  - `name`: Filter by trade name
  - `generic_name`: Filter by active ingredient
  - `batch_number`: Exact batch match
  - `status`: `active`, `expired`, `recalled`, `suspended`
  - `manufacturer_id`: UUID
  - `expiry_before`, `expiry_after`: Date filters (`YYYY-MM-DD`)

#### `GET /api/v1/drugs/:id`
Retrieves full details of a drug batch.
* **Auth**: Bearer Token

#### `PATCH /api/v1/drugs/:id`
Updates drug batch metadata or status.
* **Auth**: Bearer Token

#### `DELETE /api/v1/drugs/:id`
Deletes a drug batch if no units are linked.
* **Auth**: Bearer Token

---

### 5. Drug Units & Serialization

#### `POST /api/v1/drugs/:id/units`
Generates individual serialized units with cryptographically random serial numbers.
* **Auth**: Bearer Token
* **Request Body**:
```json
{
  "quantity": 50
}
```
* **Response `201 Created`**:
```json
{
  "success": true,
  "data": [
    {
      "id": "c1f103b4-5231-4122-bc03-62ef9761937f",
      "drug_id": "18f0ad56-34d2-432f-a9cb-6b2169123456",
      "serial_number": "4B892F1AC9083DE1",
      "status": "active",
      "created_at": "2026-09-23T07:15:00Z"
    }
  ]
}
```

#### `GET /api/v1/drugs/:id/units`
Lists serialized units belonging to a drug batch.
* **Auth**: Bearer Token
* **Query Params**: `page`, `limit`

#### `GET /api/v1/drug-units/:id/qr`
Generates a downloadable or scannable PNG QR code for a specific unit.
* **Auth**: Bearer Token
* **Query Params**: `size` (default: 256, max: 1024)
* **Response `200 OK`**: `image/png` binary content.

---

### 6. Public Verification (Core Feature)

#### `GET /api/v1/verify/:serial`
Verifies authenticity of a drug unit by its serial number or scanned QR code.
* **Auth**: Public (No authentication required)
* **Responses**:

##### Case 1: Verified (Authentic & Active)
```json
{
  "success": true,
  "data": {
    "verified": true,
    "result": "verified",
    "message": "Drug record verified. This serial number corresponds to a registered and active drug.",
    "drug": {
      "name": "Augmentin 625mg Tablets",
      "generic_name": "Amoxicillin / Clavulanic Acid",
      "manufacturer": "GlaxoSmithKline Pharmaceuticals",
      "batch_number": "AUG-2026-09",
      "nafdac_number": "A4-0123",
      "manufacturing_date": "2026-01-15",
      "expiry_date": "2028-01-15"
    },
    "drug_unit": {
      "serial_number": "4B892F1AC9083DE1",
      "status": "active"
    }
  }
}
```

##### Case 2: Counterfeit / Not Found
```json
{
  "success": true,
  "data": {
    "verified": false,
    "result": "not_found",
    "message": "The drug could not be verified. No record found for this serial number."
  }
}
```

##### Case 3: Expired Drug
```json
{
  "success": true,
  "data": {
    "verified": false,
    "result": "expired",
    "message": "This drug has expired. Do not use this product.",
    "drug": {
      "name": "Augmentin 625mg Tablets",
      "generic_name": "Amoxicillin / Clavulanic Acid",
      "manufacturer": "GlaxoSmithKline Pharmaceuticals",
      "batch_number": "AUG-2026-09",
      "nafdac_number": "A4-0123",
      "manufacturing_date": "2022-01-15",
      "expiry_date": "2024-01-15"
    },
    "drug_unit": {
      "serial_number": "4B892F1AC9083DE1",
      "status": "active"
    }
  }
}
```

##### Case 4: Recalled Drug
```json
{
  "success": true,
  "data": {
    "verified": false,
    "result": "recalled",
    "message": "This drug batch has been recalled. Do not use this product."
  }
}
```

---

### 7. Verification Audit Logs

#### `GET /api/v1/verification-logs`
Lists all public verification attempts across all drugs for audit and counterfeit detection.
* **Auth**: Bearer Token
* **Query Params**:
  - `page`, `limit`
  - `serial_number`: Search specific serial
  - `result`: Filter by `verified`, `not_found`, `expired`, `recalled`, `suspended`, `invalid`
  - `from_date`, `to_date`: RFC3339 timestamps

#### `GET /api/v1/drugs/:id/verification-logs`
Lists verification attempts for a specific drug batch.
* **Auth**: Bearer Token
* **Query Params**: Same as global verification logs.
