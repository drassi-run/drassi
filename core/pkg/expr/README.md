# GHA Expression Parser

## Supported Features:

### Literals:
- [x] String
- [x] Number
- [x] Boolean

### Functions:

#### String Formatting Functions:
- [x] `contains`
- [x] `endsWith`
- [x] `startsWith`
- [x] `join`
- [x] `format`

#### Status Check Functions:
- [x] `always`
- [x] `cancelled` (currently only supports `job.status` for pre, post and job-level steps)
- [x] `success` (currently only supports `job.status` for pre, post and job-level steps)
- [ ] `fromJSON`
- [ ] `toJson`
- [ ] `hashFiles`
