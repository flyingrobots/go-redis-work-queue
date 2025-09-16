# RBAC and Tokens Test Suite Documentation

This documentation describes the comprehensive test suite for the RBAC (Role-Based Access Control) and Tokens feature (F008) in the go-redis-work-queue system.

## Test Coverage Overview

The test suite provides comprehensive coverage across multiple layers:

- **Unit Tests**: 400+ lines covering core RBAC logic
- **Integration Tests**: 300+ lines testing API interactions
- **E2E Tests**: 600+ lines testing complete workflows
- **Security Tests**: 500+ lines testing attack vectors

**Total Test Code**: ~1,800 lines
**Target Coverage**: 80% code coverage
**Test Categories**: Unit, Integration, E2E, Security

## Test Structure

### Unit Tests (`internal/rbac-and-tokens/`)

#### `auth_test.go` - Core Authentication Logic
Tests the fundamental authentication and authorization mechanisms:

**TestTokenValidation**
- Valid token processing
- Expired token rejection
- Invalid signature detection
- Malformed token handling
- JSON payload validation

**TestTimeSkewTolerance**
- Clock drift tolerance (60-second skew)
- Future token rejection (nbf validation)
- Edge case temporal validation
- Expired token detection

**TestScopeMatching**
- Exact scope matching
- Admin all-permissions logic
- Multiple scope validation
- Empty scope handling

**TestRolePermissions**
- Role hierarchy validation
- Permission inheritance (Viewer < Operator < Maintainer < Admin)
- Role-to-permission mapping
- Admin all-access validation

**TestTokenRevocation**
- Revocation list management
- Revoked token rejection
- Revocation reason tracking
- Multiple revocation handling

#### `authorization_test.go` - Advanced Authorization Logic

**TestResourcePatternMatching**
- Wildcard pattern matching (`queue:payment-*`)
- Exact resource matching
- Complex pattern support (`*-high`, `queue:*:jobs`)
- Case-sensitive matching
- Global wildcard (`*`)

**TestScopeAuthorization**
- Direct scope authorization
- Role-based authorization
- Resource constraint enforcement
- Admin override behavior

**TestRoleHierarchy**
- Role inheritance testing
- Permission escalation prevention
- Hierarchical access control

**TestAuditLogging**
- Audit entry validation
- Structured logging format
- Required field validation
- Audit entry serialization

### Integration Tests (`test/integration/`)

#### `rbac_integration_test.go` - Full API Integration

**TestRBACIntegrationFullFlow**
Tests complete RBAC workflows with different user roles:

- **Viewer Role**: Read-only access validation
- **Operator Role**: Read/write with restricted delete
- **Maintainer Role**: Maintenance operations access
- **Admin Role**: Full system access

Each role test includes:
- Stats endpoint access
- Queue operations (peek, enqueue)
- Destructive operations (DLQ purge, worker restart)
- Permission boundary enforcement

**TestResourceConstraints**
- Resource pattern enforcement (`payment-*`, `*-high`)
- Multi-tenant access control
- Cross-tenant access prevention

**TestTokenRevocationIntegration**
- Real-time token revocation
- Post-revocation access denial
- Revocation workflow validation

**TestAuditLoggingIntegration**
- Audit trail generation
- Destructive operation logging
- Compliance requirement validation

### E2E Tests (`test/e2e/`)

#### `rbac_e2e_test.go` - Complete System Workflows

**TestE2ETokenLifecycle**
Tests realistic user workflows:

1. **DevOps Engineer Scenario**
   - System monitoring
   - Deployment job enqueuing
   - Queue status monitoring
   - Denied destructive operations

2. **Site Reliability Engineer Scenario**
   - System health checks
   - Dead letter queue management
   - Worker management
   - Boundary enforcement

3. **Security Admin Scenario**
   - Full system access
   - Emergency response capabilities
   - Performance benchmarking
   - System-wide operations

**TestE2ESecurityBoundaries**
- Token forgery detection
- Privilege escalation prevention
- Replay attack protection
- Resource boundary enforcement

**TestE2EMultiTenancy**
- Tenant isolation validation
- Cross-tenant access prevention
- Multi-tenant resource patterns

### Security Tests (`internal/admin-api/`)

#### `rbac_security_test.go` - Security Vulnerability Testing

**TestSecurityFuzzHeaders**
Fuzzing attack testing for:
- Authorization header manipulation
- Content-Type injection attempts
- User-Agent exploitation
- X-Forwarded-For spoofing
- Path traversal attempts
- XSS injection vectors
- SQL injection attempts

**TestSecurityScopeEscalation**
- Token tampering detection
- Role hierarchy bypass attempts
- Scope injection attacks
- Algorithm confusion attacks ("none" algorithm)
- Malformed claim exploitation

**TestSecurityReplayAttacks**
- Expired token replay
- Future token exploitation
- Clock skew manipulation
- Modified timestamp attacks

**TestSecurityTimingAttacks**
- Consistent timing validation
- Information leakage prevention
- Statistical timing analysis

**TestSecurityResourceExhaustion**
- Large token DoS protection
- Rate limiting validation
- Resource consumption monitoring

## Test Execution

### Running Unit Tests
```bash
go test -v ./internal/rbac-and-tokens/
```

### Running Integration Tests
```bash
go test -v ./test/integration/
```

### Running E2E Tests
```bash
go test -v ./test/e2e/
```

### Running Security Tests
```bash
go test -tags security -v ./internal/admin-api/ -run "TestSecurity"
```

### Coverage Analysis
```bash
go test -coverprofile=coverage.out ./internal/rbac-and-tokens/
go tool cover -html=coverage.out
```

## Test Scenarios by Category

### Authentication Scenarios
- [x] Valid JWT token validation
- [x] Expired token rejection
- [x] Invalid signature detection
- [x] Malformed token handling
- [x] Time skew tolerance
- [x] Token revocation enforcement

### Authorization Scenarios
- [x] Role-based permission checking
- [x] Scope-based authorization
- [x] Resource pattern matching
- [x] Admin override behavior
- [x] Permission inheritance
- [x] Access denial logging

### Security Scenarios
- [x] Header injection prevention
- [x] Token forgery detection
- [x] Privilege escalation blocking
- [x] Replay attack protection
- [x] DoS attack mitigation
- [x] Information leakage prevention

### Integration Scenarios
- [x] Full API workflow validation
- [x] Multi-tenant access control
- [x] Audit trail generation
- [x] Real-time token management
- [x] Resource constraint enforcement

## Test Data and Fixtures

### Standard Test Roles
- **Viewer**: `PermStatsRead`, `PermQueueRead`, `PermJobRead`, `PermWorkerRead`
- **Operator**: Viewer permissions + `PermQueueWrite`, `PermJobWrite`, `PermBenchRun`
- **Maintainer**: Operator permissions + `PermQueueDelete`, `PermJobDelete`, `PermWorkerManage`
- **Admin**: All permissions via `PermAdminAll`

### Test Token Formats
- Standard JWT with HMAC-SHA256
- Claims include: `sub`, `roles`, `scopes`, `exp`, `iat`, `nbf`, `iss`, `aud`, `jti`, `kid`
- Resource constraints in `resources` claim
- Token types: `bearer`, `api_key`, `session`

### Test Endpoints
- `GET /api/v1/stats` - Statistics access
- `GET /api/v1/queues/{queue}/peek` - Queue inspection
- `POST /api/v1/queues/{queue}/enqueue` - Job submission
- `DELETE /api/v1/queues/dlq` - DLQ management
- `DELETE /api/v1/queues/all` - System purge
- `POST /api/v1/bench` - Performance testing

## Security Test Payloads

### Malicious Headers
- Path traversal: `../../../etc/passwd`
- XSS injection: `<script>alert('xss')</script>`
- Header injection: `header\r\nX-Evil: true`
- SQL injection: `' OR 1=1 --`
- Null byte injection: `\x00\x01\x02\x03`

### Attack Tokens
- Unsigned tokens (algorithm: none)
- Tampered scope claims
- Modified timestamps
- Excessive payload sizes
- Malformed JSON structures

## Performance Benchmarks

### Token Validation
- Target: < 5ms p99 latency
- Throughput: 10,000 ops/sec per core
- Memory: ~50KB per 1,000 active tokens

### Authorization Checks
- Target: < 1ms p99 latency
- Cache hit ratio: > 95%
- Concurrent requests: 1,000/sec

### Audit Logging
- Async processing: 5,000 events/sec
- Storage growth: ~1KB per event
- Retention: 2 years

## Quality Gates

### Coverage Requirements
- **Unit Tests**: ≥80% line coverage
- **Integration Tests**: ≥70% scenario coverage
- **E2E Tests**: Critical path coverage
- **Security Tests**: Attack vector coverage

### Performance Requirements
- All tests complete in < 30 seconds
- No memory leaks during execution
- Resource cleanup validation
- Concurrent test execution support

### Security Requirements
- Zero information leakage
- All attack vectors blocked
- Consistent error handling
- Audit trail completeness

## Maintenance

### Test Updates Required When:
- Adding new RBAC roles or permissions
- Modifying token claim structure
- Changing API endpoints
- Updating security policies
- Adding new attack vectors

### Continuous Integration
- All tests run on every PR
- Security tests run nightly
- Performance regression detection
- Coverage reporting integration

---

This test suite ensures the RBAC and Tokens system meets enterprise security standards while maintaining high performance and reliability. The comprehensive coverage across unit, integration, E2E, and security testing provides confidence in the system's robustness against both functional and security requirements.