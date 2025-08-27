# Bookmark Sync Service API Documentation

## Overview
A comprehensive bookmark synchronization service with cross-browser support

## Base URLs
- **Development server**: http://localhost:8080/api/v1
- **Production server**: https://api.bookmark-sync.com/v1

## Authentication
This API uses JWT Bearer token authentication. Include the token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## Authentication

### POST /auth/login

Authenticate user and return JWT token

#### Request Body

User credentials

```json
{
  // Request body example
}
```

#### Responses

**401**: Invalid credentials

**200**: Login successful


---

### POST /auth/register

Register new user account

#### Request Body

User registration data

```json
{
  // Request body example
}
```

#### Responses

**201**: User created successfully

**400**: Invalid input data


---

## Bookmarks

### GET /bookmarks

List user bookmarks with pagination and filtering

#### Parameters

| Name | Type | In | Required | Description |
|------|------|----|---------| ----------- |
| limit | integer | query | No | Number of bookmarks to return |
| offset | integer | query | No | Number of bookmarks to skip |
| search | string | query | No | Search term for filtering bookmarks |

#### Responses

**200**: List of bookmarks


---

### POST /bookmarks

Create a new bookmark

#### Request Body

Bookmark data

```json
{
  // Request body example
}
```

#### Responses

**201**: Bookmark created successfully

**400**: Invalid bookmark data


---

## Collections

### GET /collections

List user collections

#### Responses

**200**: List of collections


---

### POST /collections

Create a new collection

#### Request Body

Collection data

```json
{
  // Request body example
}
```

#### Responses

**201**: Collection created successfully


---

