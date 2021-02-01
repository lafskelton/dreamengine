# dreamengine

  ### Abstract 
  ---

  - A cloud AI compute engine that serves multiple neural models concurrently. 
  - uses a standardized model stored in a centralized library, allowing containers to load any model on-the-fly
  - Managed gRPC stream is used as a medium of communication.
  - Engine is written on C cuda 11.2, uses a single gpu
  - Server is written in go for optimal concurrency and performance at scale
  
  
  ### Road map:
  ---
  
  - multi-gpu
  - distribution & redundancy
  - external library service
  - model training
  - optimizers
  - CUDA warp matrix capability (enable use of tensor cores)

  
