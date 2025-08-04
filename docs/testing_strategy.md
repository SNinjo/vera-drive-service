# Testing Strategy

## Overview

This project follows a **layered testing approach** with comprehensive unit tests and focused API/E2E tests.

## Testing Philosophy

### Unit Tests ✅ (Comprehensive)
- **Purpose**: Test each scenario and edge case
- **Scope**: Individual functions, methods, and components
- **Coverage**: Success cases, error cases, edge cases
- **Speed**: Fast execution (< 1 second)
- **Dependencies**: Mocked external dependencies

### API/E2E Tests ✅ (Focused)
- **Purpose**: Verify system integration and happy paths
- **Scope**: Complete user journeys and system boundaries
- **Coverage**: Happy path scenarios and critical error paths
- **Speed**: Slower execution (requires full system setup)
- **Dependencies**: Real database, middleware, etc.

## Test Structure

```
├── internal/
│   ├── url/
│   │   ├── service_test.go      # Unit tests for business logic
│   │   ├── handler_test.go      # Unit tests for HTTP handlers
│   │   └── repository_test.go   # Unit tests for data access
├── test/
│   └── api_test.go              # API/E2E tests
└── docs/
    └── testing_strategy.md      # This document
```

## Unit Test Guidelines

### Service Layer Tests
- Test all business logic scenarios
- Mock repository dependencies
- Test validation rules
- Test error handling

```go
func TestService_CreateURL_Success(t *testing.T) {
    // Arrange
    mockRepo := &MockRepository{}
    service := &service{repo: mockRepo}
    
    // Act
    result := service.CreateURL(...)
    
    // Assert
    assert.NoError(t, result)
    mockRepo.AssertExpectations(t)
}
```

### Handler Layer Tests
- Test HTTP request/response handling
- Test input validation
- Test error responses
- Mock service dependencies

### Repository Layer Tests
- Test database operations
- Test SQL queries
- Test data mapping
- Use test database

## API/E2E Test Guidelines

### Focus Areas
1. **Complete User Journeys**: Test full workflows
2. **System Integration**: Verify all components work together
3. **Authentication/Authorization**: Test security boundaries
4. **Error Handling**: Test critical error scenarios

### Test Structure
```go
func TestAPI_CompleteUserJourney(t *testing.T) {
    // Setup real system
    router := setupTestRouter()
    
    t.Run("Complete URL Management Flow", func(t *testing.T) {
        // 1. Create resource
        // 2. Read resource
        // 3. Update resource
        // 4. Delete resource
        // Verify each step works
    })
}
```

## Running Tests

### Unit Tests
```bash
# Run all unit tests
go test ./internal/...

# Run specific package
go test ./internal/url/

# Run with coverage
go test -cover ./internal/...
```

### API Tests
```bash
# Run API tests
go test ./test/

# Run with verbose output
go test -v ./test/
```

## Test Data Management

### Unit Tests
- Use mocks for external dependencies
- Create test data inline
- Clean up after each test

### API Tests
- Use test database
- Create test fixtures
- Clean up database after tests

## Best Practices

### 1. Test Naming
Use descriptive names: `TestFunction_Scenario_ExpectedResult`

### 2. Arrange-Act-Assert Pattern
```go
func TestSomething(t *testing.T) {
    // Arrange - Setup test data and mocks
    mockRepo := &MockRepository{}
    
    // Act - Execute the function being tested
    result := function(mockRepo)
    
    // Assert - Verify the results
    assert.Equal(t, expected, result)
}
```

### 3. Table-Driven Tests
For multiple similar test cases:
```go
func TestValidateInput(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected error
    }{
        {"valid input", "valid", nil},
        {"empty input", "", ErrEmptyInput},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validateInput(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 4. Mock Expectations
Always verify mock expectations:
```go
mockRepo.On("Create", mock.Anything).Return(nil)
// ... test code ...
mockRepo.AssertExpectations(t)
```

## Coverage Goals

- **Unit Tests**: 90%+ code coverage
- **API Tests**: Critical user journeys covered
- **Integration**: Database and external service integration

## Continuous Integration

- Run unit tests on every commit
- Run API tests on pull requests
- Generate coverage reports
- Fail builds on test failures

## Why This Approach?

1. **Unit Tests**: Fast feedback, comprehensive coverage, easy debugging
2. **API Tests**: Confidence in system integration, catch integration bugs
3. **Separation of Concerns**: Unit tests focus on logic, API tests focus on integration
4. **Maintainability**: Clear test boundaries, easy to understand and modify 