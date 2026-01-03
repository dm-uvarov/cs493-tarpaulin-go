# Tarpaulin REST API

## Overview

Tarpaulin is a cloud-deployed RESTful backend service for managing course data with secure, role-based access control.  
The API supports authenticated users with different roles (admin, instructor, student) using JWT-based authentication.

The project focuses on backend API design, authorization, and cloud deployment.

---

## Features

- RESTful API built with Go and the Gin framework
- JWT-based authentication
- Role-Based Access Control (RBAC)
  - Admin
  - Instructor
  - Student
- CRUD operations for course-related resources
- File storage using Google Cloud Storage
- Deployed on Google App Engine
- API tested with Postman and Newman

---

## Tech Stack

- **Language:** Go  
- **Framework:** Gin  
- **Authentication:** JWT  
- **Authorization:** Role-Based Access Control (RBAC)  
- **Database:** Google Cloud Datastore (NoSQL)  
- **Cloud Platform:** Google App Engine  
- **Storage:** Google Cloud Storage  
- **Testing:** Postman, Newman  

---

## API Functionality

Access to endpoints is restricted based on user role:

### Admin
- Create and manage users
- Full access to all resources

### Instructor
- Create and manage courses
- Upload and manage course-related files

### Student
- View course information
- Access permitted course resources

Unauthorized requests return appropriate HTTP status codes.

---

## Authentication & Authorization

- Authentication is handled using JSON Web Tokens (JWT)
- Tokens must be included in request headers
- Role permissions are validated on every protected endpoint

---

## Deployment

The application is deployed using Google App Engine and integrates with:
- Google Cloud Datastore for persistent data
- Google Cloud Storage for file uploads

---

## Testing

API endpoints were tested using:
- Postman for manual testing
- Newman for automated API testing

Testing covered:
- Authentication flows
- Role-based access control
- Error handling and response validation

---

## Project Context

This project was developed as a backend-focused coursework assignment emphasizing:
- REST API design
- Secure authentication
- Authorization
- Cloud deployment

The repository is intended to demonstrate backend engineering ownership.

---

## Notes

This is an educational project and does not include a frontend interface.
