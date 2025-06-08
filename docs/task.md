You will implement a complete RESTful API for an application called Tarpaulin, a lightweight course management tool that's an "alternative" to Canvas. You can find more details about Tarpaulin below. The API you implement will utilize most of the components of a modern API that we have talked about in this class. You will deploy your application on Google Cloud Platform using Google App Engine and Datastore, using go language. You will use Auth0 for authentication.

# Required Functionality

The application for which you’ll write a REST API for this project is **Tarpaulin**, a lightweight course management tool. The Tarpaulin REST API has **13 endpoints**, most of which are protected. The protected endpoints require a valid JWT in the request as a Bearer token in the Authorization header. Each user in Tarpaulin has one of three roles: **admin, instructor, and student**.

Below is a summary of the endpoints you need to implement (the link to the full API spec is provided in the next section):

| #  | Functionality                         | Endpoint                        | Protection                                | Description                                                                                      |
|----|---------------------------------------|---------------------------------|-------------------------------------------|--------------------------------------------------------------------------------------------------|
| 1  | User login                           | POST /users/login               | Pre-created Auth0 users                   | Use Auth0 to issue JWTs. You may use the example app code from "Exploration - Implementing Auth Using JWTs" with minor response changes. |
| 2  | Get all users                        | GET /users                      | Admin only                                | Summary info of all 9 users. No info about avatar or courses.                                    |
| 3  | Get a user                           | GET /users/:id                  | Admin or user with matching JWT           | Detailed user info, including avatar (if any) and courses (for instructors and students).        |
| 4  | Create/update a user’s avatar         | POST /users/:id/avatar          | User with matching JWT                    | Upload file to Google Cloud Storage.                                                            |
| 5  | Get a user’s avatar                   | GET /users/:id/avatar           | User with matching JWT                    | Read and return file from Google Cloud Storage.                                                 |
| 6  | Delete a user’s avatar                | DELETE /users/:id/avatar        | User with matching JWT                    | Delete file from Google Cloud Storage.                                                          |
| 7  | Create a course                       | POST /courses                   | Admin only                                | Create a course.                                                                                |
| 8  | Get all courses                       | GET /courses                    | Unprotected                               | Paginated using offset/limit. Page size is 3. Ordered by "subject." No course enrollment info.   |
| 9  | Get a course                          | GET /courses/:id                | Unprotected                               | No course enrollment info returned.                                                             |
| 10 | Update a course                       | PATCH /courses/:id              | Admin only                                | Partial update.                                                                                 |
| 11 | Delete a course                       | DELETE /courses/:id             | Admin only                                | Delete course and related enrollment info.                                                      |
| 12 | Update enrollment in a course         | PATCH /courses/:id/students     | Admin or instructor of the course         | Enroll or disenroll students from the course.                                                   |
| 13 | Get enrollment for a course           | GET /courses/:id/students       | Admin or instructor of the course         | All students enrolled in the course.                                                            |


REST API Specification Document

Here is the doc for the REST API  you need to implement.(docs/assignment6-api-doc.pdf) This API explains the functionality outlined above. The REST API doc includes

Details of the endpoints (i.e., URL + method combos) you need to implement
Details of the format of requests and responses for each endpoint
Errors your code needs to catch, and
Status codes that must be returned in case of success and in case of errors.
Read this document carefully because it explains in detail the required implementation of all the endpoints in the REST API.


Data

There are several kinds of data Tarpaulin will need to keep track of.

Users

These represent Tarpaulin application users.

Each user has one of three roles: admin, instructor, and student.
Each of these roles represents a different set of permissions to perform certain API actions.
Many of the endpoints in the Tarpaulin API require authorization, as described in the Tarpaulin specification.
Create 9 Users in Auth0
You must use Auth0's JWT-based authorization scheme to issue JWTs for the users.
For this, you must pre-create the following 9 users in Auth0:
A user with username admin1@osu.com.
This user will have the role admin in Datastore (as described below)
Two users with usernames instructor1@osu.com and instructor2@osu.com
These users will have the role instructor in Datastore (as described below).
Six users with usernames student1@osu.com, student2@osu.com, and so on until student6@osu.com.
These users will have the role student in Datastore (as described below).
All 9 users must have the same password. You can choose whatever password you want.
Note that we don't store any info in Auth0 about a user's role. Info about user roles is stored only in Datastore. 
Populate Data for the 9 Users in Datastore
You must pre-populate some data for these users in Datastore. The following video describes how you can start populating this data.

You must create a kind users in Datastore that has the following 3 properties:
id: integer generated by Datastore.
sub: The value of sub assigned to this user by Auth0.
role: The role of the user.
You must pre-populate the kind users with the 9 users listed above
The kind users will have exactly these 9 users (i.e., don't add any other users in the kind users).
The pre-populated data will only have values for the 3 properties, id, sub and role.
You can design the kind users to have additional properties whose values are set, modified, deleted, etc., by the REST API. But the REST API must not modify these 3 pre-set properties of the 9 users.
The username and password of these 9 users must be stored in Auth0 and not in Datastore.
Once you have implemented the User Login functionality of your REST API, i.e., the endpoint POST /users/login, you can use the Postman Collection we have provided to populate the sub property of the 9 users in the kind users. This is described in the following video.

Courses & Enrollment

These represent courses being managed in Tarpaulin and the enrollment of students in these courses.
Each course has basic information, such as subject code, number, title, instructor, etc.
Each course also has associated data to keep track of students (i.e. Tarpaulin users with the student role) who are enrolled in the course.
It is up to you to design the data model to store the data about courses and enrollment (i.e., what kind or kinds you will create, and what will be the properties of these kinds).

The assignment requires using some Datastore features (multiple conditions, ordering results, limit-offset based pagination) that we haven't used before in the class. We have provided you notes on these features  that will help you implement the functionality.(docs/advanced datastore queries.pdf) in python 3
You cannot store the values of any URLs (e.g., self, avatar URL, etc.) in the database.
You must use Google Cloud Storage to store the image files for the users' avatar.
All functionality for endpoints 2 through 13 must be implemented by performing CRUD operations on data stored in Datastore and operations on Google Cloud Storage.
This means that you cannot hard-code any data from the kind users in your program, or cache this data in memory.
Testing the API

Note: Unlike previous assignments, the Postman Collection provided to you does not include tests for the complete API you will implement. Be sure to review information, including videos, in this section to understand which parts of the API are not covered by the Postman Collection we have provided.

You can test most parts of your API by downloading the following two JSON files, one file has a Postman Collection and the other has a Postman Environment.

Postman Collection Download Postman Collection
These tests cover the functionality your API must implement for endpoints 1, 2, 4, 5, 6, 7, 8, 9.
Tests for endpoint 4 (create/update a user's avatar) require some updates to set the value of the image file to upload.
Tests for endpoint 5 (get a user's avatar) don't validate that the correct file is returned. This criterion will be manually graded.
For endpoint 3 (get a user), the collection includes tests that cover some basic functionality for the endpoint. However, the functionality related to the value of a user's property "courses" is not tested by the collection.
We are not providing tests for endpoints 10, 11, 12 and 13.
We recommend that you write your own Postman tests to test these endpoints, and to test the part of the functionality of endpoint 3 not covered by the test collection we have provided.
If you write tests for this functionality, you don't need to submit those tests or any Postman collection. We will test these endpoints using our own grading tests that we will not share with the students.
The Postman Collection file has logging statements which will show you how many points you received from running the collection (similar to what is shown in the video for Assignment 2).
Postman Environment Download Postman Environment
The environment file contains certain pre-defined variables used by the tests.
Don't change any pre-defined variables, except the variables app_url and password.
You must set the correct value of these 2 variables in the Postman Environment file to run the Postman collection.
app_url
This variable allows you to run the test suite against the service running on your computer or running it against the service deployed on GCP.
To run against service on your computer, set it to the URL of the service running on your computer, e.g., http://localhost:8080.
To run against service deployed on GCP, set it to the URL your project is deployed to, e.g., https://my-project.appspot.com.
password
Set the value of this variable to the common password value you set for all 9 users pre-created in Auth0 for this assignment.
Tests for Endpoints 2 and 3

The following video highlights some points about the tests provided for endpoint 2, GET /users, and endpoint 3, GET /users/:id.
Tests for Endpoints 3b, 4 and 5

The following video highlights some points about the tests provided for the functionality of

endpoint 3, GET /users/:id, related to avatars
endpoint 4, POST /users/:id/avatar, and
endpoint 5, GET /users/:id/avatar.