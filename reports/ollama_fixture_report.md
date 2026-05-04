# ProtoGuard Ollama reports (testdata fixtures)

Generated: 2026-05-04T06:30Z

---

## OpenAPI: testdata/openapi_old.yaml → openapi_new.yaml

# ProtoGuard diff (openapi)

Changes: **2** — includes **BREAKING**

| Impact | Count |
|---|---:|
| BREAKING | 1 |
| NON_BREAKING | 1 |

## Structured changes

- `openapi.paths./cats` **add** (NON_BREAKING): added path
- `openapi.paths./pets` **remove** (BREAKING): removed path

## AI explanations

### Chunk 1

**API Compatibility Analysis Report**
=====================================

### Summary

This report summarizes the changes made to the API specification from version X to version Y.

### Breaking Changes

None detected in this analysis.

### Non-breaking changes

* Added a new path `/cats` under `openapi.paths./cats`. This change is marked as non-breaking, indicating that it does not affect existing functionality.
	+ Impact: NON_BREAKING
	+ Summary: added path

### Risky changes

None detected in this analysis.

### Suggested backward-compatible fixes

* To ensure compatibility with older clients, consider adding a `deprecated` keyword to the new `/cats` endpoint, indicating that it will be removed in a future version.
* Consider adding a `x-protobuf-field-reservation` header to the response of the new `/cats` endpoint, reserving a specific protobuf field for future use.

Note: These suggestions are based on general best practices and may not be applicable to all scenarios. It's essential to review the specific requirements and constraints of your project before implementing any changes.

### Chunk 2

**Summary**
------------

* API version changed to OpenAPI.

**Breaking changes**
------------------

### Removed Path

* `"/pets"` has been removed from the API.
* This change is marked as BREAKING, indicating that it may cause issues with existing clients or integrations.

No additional breaking changes were found in this JSON array.


---

## Protobuf: testdata/proto_old.pb → proto_new.pb

# ProtoGuard diff (protobuf)

Changes: **2** — includes **BREAKING**

| Impact | Count |
|---|---:|
| BREAKING | 1 |
| NON_BREAKING | 1 |

## Structured changes

- `grpc.field.pgtest.Item.1` **remove** (BREAKING): field removed (reserve number for compatibility)
- `grpc.field.pgtest.Item.2` **add** (NON_BREAKING): field added

## AI explanations

### Chunk 1

**Summary**
------------

The provided JSON array contains two structured changes for a protobuf API. The changes are related to the `grpc.field.pgtest.Item` message.

**Breaking Changes**
------------------

*   **Removed Field**: The field `title` has been removed from the `grpc.field.pgtest.Item` message. This change is marked as BREAKING, indicating that it may cause compatibility issues with existing clients or codebases.
    *   **Impact**: BREAKING
    *   **Reason**: The field was removed to reserve its position for future compatibility.

**Non-Breaking Changes**
----------------------

*   **Added Field**: A new field `code` has been added to the `grpc.field.pgtest.Item` message. This change is marked as NON_BREAKING, indicating that it will not cause any issues with existing clients or codebases.
    *   **Impact**: NON_BREAKING
    *   **Reason**: The new field was added to provide additional functionality without breaking existing compatibility.

**Risky Changes**
----------------

None of the changes in this JSON array are marked as RISKY. However, it's essential to review and test any changes to ensure they do not introduce unexpected behavior or security vulnerabilities.

**Suggested Backward-Compatible Fixes (protobuf Field Reservations)**
-------------------------------------------------------------------

To maintain compatibility with existing clients or codebases that rely on the removed `title` field:

*   Reserve its position by using a default value for the new `code` field. This can be achieved by adding a default value to the `code` field in the protobuf definition.

Example:
```protobuf
message Item {
  int32 code = 2;
  // ...
}
```
By reserving the position of the removed field, you ensure that existing clients or codebases will not break when encountering this change.

