# Project Status & Todo

## 1. Authentication & Identity
- [x] **Auth Core**: Implement Register and Login handlers.
- [x] **Token Management**: Setup JWT issuance and validation middleware.
- [x] **Security**: Implement password hashing (bcrypt/argon2).

## 2. User Management
- [x] **Create**: (Covered in Auth).
- [x] **Read**: Fetch user profile/details.
- [x] **Update**: Allow users to edit profile information.
- [x] **Delete**: Implement account deletion logic.

## 3. Library Operations
- [x] **Create**: Logic to initialize a new library.
- [x] **Read**: List available libraries and view library details.
- [x] **Update**: Modify library metadata (name, description).
- [x] **Delete**: Remove a library (and handle cascading logic).

## 4. Book Management
- [ ] **Add Book**: Schema validation and database insertion for book metadata.
- [ ] **Read**: Get single book details and paginated list of books.
- [ ] **Update**: Edit book metadata (authors, genres, year).
- [ ] **Delete**: Remove book reference.

## 5. Volume Management (File Handling)
- [ ] **Volume CRUD**:
    - [ ] Create volume record (linked to Book).
    - [ ] Retrieve volume details.
    - [ ] Update volume attributes.
    - [ ] Delete volume record.
- [ ] **Upload Flow**:
    - [ ] Initialize upload session.
    - [ ] Handle file binary/streaming.
    - [ ] Finalize upload and update storage path in DB.

## 6. Backlog / Upcoming
- [ ] Integration Testing
- [ ] API Documentation (Swagger/OpenAPI)