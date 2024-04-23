GHA expression parser.

Supported features:
  Literals:
      - [x] string, number, boolean
  
  Functions:
    String fmt functions:
      - [x] contains
      - [x] endsWith
      - [x] starsWith
      - [x] join
      - [x] format
    Status check functions:
    
      - [x] always
      - [x] cancelled (currently only support job.status' for pre, post and job-getLevel steps)
      - [x] success (currently only support job.status' for pre, post and job-getLevel steps)
      - [ ] fromJson
      - [ ] toJson
      - [ ] hashFile
