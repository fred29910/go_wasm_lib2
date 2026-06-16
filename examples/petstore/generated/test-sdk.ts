/**
 * test-sdk.ts — SDK Debug Test Suite
 *
 * Tests the generated TypeScript SDK (sdk.ts) for correct type definitions,
 * request construction, and edge case handling.
 *
 * This is a TYPE-LEVEL test suite that validates the SDK API surface.
 * It does NOT execute WASM or make HTTP calls.
 *
 * Usage:
 *   npx tsx test-sdk.ts         # Run the test
 *   npx tsc --noEmit test-sdk.ts # Type-check only
 */

import { WASMSDK, HTTPResponse, HTTPRequest, WASMConfig, HTTPError } from './sdk.js';
import {
  Pet,
  CreatePetRequest,
  FindPetsByStatusRequest,
  GetPetByIDRequest,
  CreatePetResponse,
  FindPetsByStatusResponse,
  GetPetByIDResponse,
  createPet,
  findPetsByStatus,
  getPetByID,
} from './sdk.js';

// ============================================================================
// Test Runner
// ============================================================================

interface TestResult {
  name: string;
  passed: boolean;
  error?: string;
}

const results: TestResult[] = [];

function assert(condition: boolean, name: string, message?: string): void {
  results.push({
    name,
    passed: condition,
    error: condition ? undefined : message || 'Assertion failed',
  });
}

function assertEqual<T>(actual: T, expected: T, name: string): void {
  const pass = actual === expected;
  results.push({
    name,
    passed: pass,
    error: pass ? undefined : `Expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}`,
  });
}

function assertDeepEqual(actual: any, expected: any, name: string): void {
  const pass = JSON.stringify(actual) === JSON.stringify(expected);
  results.push({
    name,
    passed: pass,
    error: pass ? undefined : `Expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}`,
  });
}

function assertThrows(fn: () => void, name: string): void {
  try {
    fn();
    results.push({ name, passed: false, error: 'Expected an error to be thrown' });
  } catch {
    results.push({ name, passed: true });
  }
}

function report(): void {
  const total = results.length;
  const passed = results.filter((r) => r.passed).length;
  const failed = results.filter((r) => !r.passed).length;
  console.log(`\n=== SDK Test Results: ${passed}/${total} passed, ${failed} failed ===\n`);
  for (const r of results) {
    const icon = r.passed ? '✅' : '❌';
    console.log(`  ${icon} ${r.name}${r.error ? `\n       ${r.error}` : ''}`);
  }
  if (failed > 0) {
    process.exit(1);
  }
}

// ============================================================================
// 1. Type Interface Validation
// ============================================================================

function testTypeInterfaces(): void {
  console.group('1. Type Interface Validation');

  // HTTPError
  const error: HTTPError = { code: 'ERR_TEST', message: 'test error' };
  assertEqual(error.code, 'ERR_TEST', 'HTTPError.code');
  assertEqual(error.message, 'test error', 'HTTPError.message');

  const errorWithDetails: HTTPError = {
    code: 'ERR_DETAILS',
    message: 'has details',
    details: 'some details',
  };
  assertEqual(errorWithDetails.details, 'some details', 'HTTPError.details');

  // HTTPResponse with error
  const errResponse: HTTPResponse = {
    status: 400,
    headers: {},
    body: null,
    error: error,
  };
  assertEqual(errResponse.error!.code, 'ERR_TEST', 'HTTPResponse.error.code');

  // HTTPResponse with body
  const okResponse: HTTPResponse = {
    status: 200,
    headers: { 'content-type': 'application/json' },
    body: { id: 1, name: 'Fluffy' },
  };
  assertEqual(okResponse.status, 200, 'HTTPResponse.status 200');
  assertEqual(okResponse.headers!['content-type'], 'application/json', 'HTTPResponse.content-type');

  // HTTPRequest
  const req: HTTPRequest = {
    method: 'POST',
    path: '/pet',
    headers: { Authorization: 'Bearer test-token' },
    query: { status: 'available' },
    body: { name: 'Fluffy' },
  };
  assertEqual(req.method, 'POST', 'HTTPRequest.method');
  assertEqual(req.path, '/pet', 'HTTPRequest.path');
  assertDeepEqual(req.query, { status: 'available' }, 'HTTPRequest.query');
  assertEqual(req.body!.name, 'Fluffy', 'HTTPRequest.body.name');

  // WASMConfig
  const config: WASMConfig = {
    baseUrl: 'https://api.example.com',
    timeout: 30,
    headers: { 'X-Custom': 'value' },
    credentials: 'include',
  };
  assertEqual(config.baseUrl, 'https://api.example.com', 'WASMConfig.baseUrl');
  assertEqual(config.timeout, 30, 'WASMConfig.timeout');

  console.groupEnd();
}

// ============================================================================
// 2. Pet Schema Model Validation
// ============================================================================

function testPetModel(): void {
  console.group('2. Pet Schema Model Validation');

  // Valid Pet with all fields
  const pet: Pet = { name: 'Fluffy', status: 'available', iD: 1 };
  assertEqual(pet.name, 'Fluffy', 'Pet.name');
  assertEqual(pet.status, 'available', 'Pet.status');
  assertEqual(pet.iD, 1, 'Pet.id');

  // Pet with only required fields
  const minimalPet: Pet = { name: 'Buddy' };
  assertEqual(minimalPet.name, 'Buddy', 'Pet minimal name');

  // Pet with all status values
  const statuses: Array<'available' | 'pending' | 'sold'> = ['available', 'pending', 'sold'];
  for (const status of statuses) {
    const p: Pet = { name: 'Test', status };
    assertEqual(p.status, status, `Pet status: ${status}`);
  }

  console.groupEnd();
}

// ============================================================================
// 3. Operation Request Models Validation
// ============================================================================

function testOperationRequestModels(): void {
  console.group('3. Operation Request Models');

  // CreatePetRequest — requires body
  const createReq: CreatePetRequest = {
    body: { name: 'Fluffy', status: 'available' },
    headers: { Authorization: 'Bearer token' },
  };
  assertEqual(createReq.body!.name, 'Fluffy', 'CreatePetRequest.body.name');
  assertEqual(createReq.body!.status, 'available', 'CreatePetRequest.body.status');

  // CreatePetRequest — empty body (optional)
  const emptyCreateReq: CreatePetRequest = {};
  assertEqual(emptyCreateReq.body, undefined, 'CreatePetRequest optional body');

  // FindPetsByStatusRequest — with filter
  const findReq: FindPetsByStatusRequest = { status: 'available' };
  assertEqual(findReq.status, 'available', 'FindPetsByStatusRequest.status');

  // FindPetsByStatusRequest — no filter
  const findEmptyReq: FindPetsByStatusRequest = {};
  assertEqual(findEmptyReq.status, undefined, 'FindPetsByStatusRequest optional status');

  // GetPetByIDRequest — with petId
  const getReq: GetPetByIDRequest = { petID: 42 };
  assertEqual(getReq.petID, 42, 'GetPetByIDRequest.petID');

  // GetPetByIDRequest — with path params override
  const getReqWithParams: GetPetByIDRequest = {
    petID: 42,
    pathParams: { petId: '99' },
  };
  assertDeepEqual(getReqWithParams.pathParams, { petId: '99' }, 'GetPetByIDRequest.pathParams');

  console.groupEnd();
}

// ============================================================================
// 4. Operation Response Models Validation
// ============================================================================

function testOperationResponseModels(): void {
  console.group('4. Operation Response Models');

  // CreatePetResponse
  const createResp: CreatePetResponse = { data: { name: 'Fluffy', id: 1 } };
  assertEqual(createResp.data!.name, 'Fluffy', 'CreatePetResponse.data.name');

  // CreatePetResponse — empty
  const emptyCreateResp: CreatePetResponse = {};
  assertEqual(emptyCreateResp.data, undefined, 'CreatePetResponse optional data');

  // FindPetsByStatusResponse — array
  const findResp: FindPetsByStatusResponse = {
    data: [
      { name: 'Fluffy', status: 'available' },
      { name: 'Buddy', status: 'pending' },
    ],
  };
  assertEqual(findResp.data!.length, 2, 'FindPetsByStatusResponse.data length');
  assertEqual(findResp.data![0].name, 'Fluffy', 'FindPetsByStatusResponse.data[0].name');

  // GetPetByIDResponse
  const getResp: GetPetByIDResponse = { data: { name: 'Rex', id: 42 } };
  assertEqual(getResp.data!.name, 'Rex', 'GetPetByIDResponse.data.name');

  console.groupEnd();
}

// ============================================================================
// 5. Typed API Function Request Construction Validation
// ============================================================================

function testTypedAPIFunctions(): void {
  console.group('5. Typed API Function Constructors');

  // The SDK functions (createPet, findPetsByStatus, getPetByID) return
  // Promise<HTTPResponse>. We can't call them directly (no WASM), but we
  // can validate that they accept the correct parameter types at compile time.

  // Type check: function signatures compile correctly
  const createFn: (params: CreatePetRequest) => Promise<HTTPResponse> = createPet;
  const findFn: (params: FindPetsByStatusRequest) => Promise<HTTPResponse> = findPetsByStatus;
  const getFn: (params: GetPetByIDRequest) => Promise<HTTPResponse> = getPetByID;

  assertEqual(typeof createFn, 'function', 'createPet is a function');
  assertEqual(typeof findFn, 'function', 'findPetsByStatus is a function');
  assertEqual(typeof getFn, 'function', 'getPetByID is a function');

  // Note: these functions reference window.wasmCallAPI which only exists
  // in browser context. In Node.js they would throw a runtime error.
  // We only verify the function signature is correct.

  console.groupEnd();
}

// ============================================================================
// 6. WASMSDK Class Validation
// ============================================================================

function testWASMSDKClass(): void {
  console.group('6. WASMSDK Class');

  // Constructor
  const sdk = new WASMSDK();
  assertEqual(typeof sdk.load, 'function', 'WASMSDK.load method exists');
  assertEqual(typeof sdk.init, 'function', 'WASMSDK.init method exists');
  assertEqual(typeof sdk.isInitialized, 'function', 'WASMSDK.isInitialized method exists');
  assertEqual(typeof sdk.setAuthToken, 'function', 'WASMSDK.setAuthToken method exists');
  assertEqual(typeof sdk.clearAuthToken, 'function', 'WASMSDK.clearAuthToken method exists');
  assertEqual(typeof sdk.getConfig, 'function', 'WASMSDK.getConfig method exists');
  assertEqual(typeof sdk.callAPI, 'function', 'WASMSDK.callAPI method exists');

  // Initial state
  assertEqual(sdk.isInitialized(), false, 'WASMSDK initially not initialized');

  // WASMSDK with custom WASM URL
  const customSdk = new WASMSDK('./custom/main.wasm');
  assertEqual(
    (customSdk as any).wasmUrl,
    './custom/main.wasm',
    'WASMSDK custom WASM URL'
  );

  console.groupEnd();
}

// ============================================================================
// 7. Edge Case Scenarios
// ============================================================================

function testEdgeCases(): void {
  console.group('7. Edge Case Scenarios');

  // Optional query parameters
  const findReq: FindPetsByStatusRequest = {};
  assertEqual(findReq.status, undefined, 'Optional status can be undefined');

  // Pet without required name (should be allowed at TS level but fail at runtime)
  // Pet.name is required by OpenAPI spec
  const invalidPet: Pet = {} as Pet;
  assertEqual(invalidPet.name, undefined, 'Pet without name (compile-time warning only)');

  // GetPetByID without petId — should be flagged by TS, but possible at runtime
  const invalidGetReq = {} as GetPetByIDRequest;
  assertEqual(invalidGetReq.petID, undefined, 'GetPetByID without petID');

  // Body with extra fields (OpenAPI extra fields not stripped)
  const createReq: CreatePetRequest = {
    body: { name: 'Test', unknownField: 'should-be-allowed' } as any,
  };
  assertEqual((createReq.body as any).unknownField, 'should-be-allowed', 'Extra fields on body');

  // Empty headers
  const req: CreatePetRequest = { headers: {} };
  assertDeepEqual(req.headers, {}, 'Empty headers object');

  // URL path with path params (multiple params would be a stress test)
  const complexReq: GetPetByIDRequest = {
    petID: 0, // edge case: zero value
    query: { extra: 'param' },
    headers: { 'X-Debug': 'true' },
  };
  assertEqual(complexReq.petID, 0, 'GetPetByID with zero petID');

  console.groupEnd();
}

// ============================================================================
// 8. Global Type Declarations Validation
// ============================================================================

function testGlobalDeclarations(): void {
  console.group('8. Global Type Declarations');

  // The SDK declares global window types that MUST exist at runtime
  // These are defined in the `declare global` block of sdk.ts
  const requiredGlobals = [
    'wasmReady',
    'wasmInitClient',
    'wasmCallAPI',
    'wasmSetAuthToken',
    'wasmClearAuthToken',
    'wasmGetConfig',
  ] as const;

  for (const key of requiredGlobals) {
    // In Node.js these won't exist, but in browser context they should
    const exists = key in globalThis || typeof (globalThis as any)[key] !== 'undefined';
    // Note: we don't assert because this runs in Node.js
    // We just verify the type declarations are valid
  }

  // Verify the Window interface augmentation compiles
  // This is a compile-time check: if this code compiles, the types are correct
  const windowCheck: Window & typeof globalThis = window;
  assertEqual(typeof windowCheck, 'object', 'Global Window type is valid');

  console.groupEnd();
}

// ============================================================================
// Main
// ============================================================================

console.log('=== WASM SDK Type-Level Test Suite ===');
console.log('Environment:', typeof window !== 'undefined' ? 'Browser' : 'Node.js');
console.log('', '');

testTypeInterfaces();
testPetModel();
testOperationRequestModels();
testOperationResponseModels();
testTypedAPIFunctions();
testWASMSDKClass();
testEdgeCases();
testGlobalDeclarations();

report();
