# Feature Implementation Prompt Template

## Task
Build a REST API for a todo list application.

## Requirements
- [ ] Create Todo model with: id, title, description, completed, createdAt
- [ ] Implement CRUD endpoints:
  - GET /todos - List all todos
  - GET /todos/:id - Get single todo
  - POST /todos - Create todo
  - PUT /todos/:id - Update todo
  - DELETE /todos/:id - Delete todo
- [ ] Add input validation
- [ ] Write tests with >80% coverage
- [ ] Handle errors gracefully

## Process
1. Set up project structure
2. Implement data model
3. Build API endpoints one by one
4. Add validation
5. Write tests
6. Run tests and fix any failures
7. Repeat steps 5-6 until all tests pass

## Success Criteria
When ALL of the following are true:
- All CRUD endpoints implemented and working
- Input validation in place
- All tests passing
- Coverage meets threshold

## Completion
When complete, output exactly: **COMPLETE**

If stuck after 10 iterations without progress:
- Document what's blocking
- List attempted solutions
- Suggest alternative approaches
