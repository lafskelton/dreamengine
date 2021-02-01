#include <stdio.h>
#include <cuda.h>
#include <cuda_profiler_api.h>
#include <time.h>
//CLASSES
//Layer Class
//
#define imin(a,b) (a<b?a:b)
#define ieven(a) ((a%2)==0?true:false)
#define errorChk(a) (a!=cudaSuccess?a:cudaSuccess)

#define imax(a,b) (a>b?a:b)

#define BLOCK_SIZE 16

#define gpuErrchk(ans) { gpuAssert((ans), __FILE__, __LINE__); }
inline void gpuAssert(cudaError_t code, const char *file, int line, bool abort=true)
{
   if (code != cudaSuccess) 
   {
      fprintf(stderr,"GPUassert: %s %s %d\n", cudaGetErrorString(code), file, line);
      if (abort) exit(code);
   }
}
//
///
//
typedef struct {
  int layer_index; 
  int neurons;
  int prev_neurons;
  bool ready;
  int max;
  //
  float*  cuda_weights; //Temporary vector for loading values to GPU
  float*  cuda_biases;  //Temporary vector for loading values to GPU
  float** weights_by_neuron;     //2D array of sets of weights indexed for each neuron, this makes parallel exec easier
} LayerDefinition;
//Network Class
//
class ModelDefinition
{
private:
  //
  unsigned long model_id;  //lamlearn cloud ID 
  unsigned int layer_num;
  int append_index;
  LayerDefinition* layers;

public: 
  //
  //Constructor
  ModelDefinition(unsigned int model_id,unsigned int layer_num);
  void set_values(unsigned int model_id,unsigned int layer_num);
  void init_layers();
  //
  //Setters
  void append_layer(LayerDefinition layerDef);
  //functions
  //
  //Getters
  unsigned int get_id(){return model_id;};
  int get_append_index(){return append_index;};
  LayerDefinition* get_layer(int id){return &layers[id];}
  LayerDefinition* get_layer_arr_ptr(){return layers;}
  unsigned int get_layer_num(){return layer_num;};
  //
  //dispose
  //
  void dispose(){
    cudaFree(layers);
  }
};
//Constructor function
ModelDefinition::ModelDefinition(unsigned int model_id, unsigned int layer_num){
  //
  set_values(model_id, layer_num);
} 
void ModelDefinition::set_values(unsigned int input_model_id, unsigned int input_layer_num){
  //
  model_id  = input_model_id;
  layer_num = input_layer_num;
}
//ModelDefinition Methods
void ModelDefinition::init_layers(){
  //
  // Malloc in managed memory
  cudaMallocManaged((void**)&layers, layer_num * sizeof(LayerDefinition));
}
//
void ModelDefinition::append_layer(LayerDefinition layerDef){
  //
  layers[append_index] = layerDef;
  append_index++;
  return;
};
//
//
// device functions
//
__device__ float ReLU(float output){
  //
  if(output < 0){
    return 0;
  } 
  return output;
}
//
__device__ float Sigmoid(float output){
  //
  return output;
}
//
__device__ float Softmax(float output){
  //
  return output;
}
//


__global__ void fast_layer_forward(int N, float* inputs, float *output_buffer, float* weights, float* biases){
  //
  int j = blockIdx.x * blockDim.x + threadIdx.x; // i/max = neuron
  int i = blockIdx.y * blockDim.y + threadIdx.y; // j     = vec pos
  //
  if(j == 0){
    output_buffer[i] = 0.0f;
  } 
  __syncthreads();
  output_buffer[i] += weights[i * N + j]*inputs[j];
  __syncthreads();
  if(j == N-1){
      output_buffer[i] = ReLU(output_buffer[i]+biases[i]);
  }
  //
  return;
}
//
__global__ void fast_4d_forward(int prevN, float* inputs, float *output_buffer, float* weights, float* biases){
  //
  if(threadIdx.x == 0){
    output_buffer[blockDim.x] = 0.0f;
  } 
  __syncthreads();
  output_buffer[blockIdx.x] += weights[(gridDim.x*prevN)+threadIdx.x]*inputs[threadIdx.x];
  __syncthreads();
  if(blockDim.x == threadIdx.x+1){
    output_buffer[blockIdx.x] = ReLU(output_buffer[blockIdx.x]+biases[blockIdx.x]);
  }
  //
  return;
}


void reduce_fraction (int &num, int &den){
  //
  for (int i = den * num; i > 1; i--) {  
    //
    if ((den % i == 0) && (num % i == 0)){ 
      //
      den /= i;  
      num /= i;  
    }      
  }
}

//
//Exported to cGO 
extern "C" {

  int get_max_threads(int device){
    cudaDeviceProp prop;
    cudaGetDeviceProperties(&prop, device);
    return prop.maxThreadsPerBlock;
  }
  char* get_device_name(int device){
    cudaDeviceProp prop;
    cudaGetDeviceProperties(&prop, device);
    char *name = (char*)prop.name;
    return name;
  }
  //
  __host__ void execute(int max_threads, char* ptr, float* in, float *out){
    //

    //
    //
    ModelDefinition * model = (ModelDefinition*)ptr;
    int num_layers = model->get_layer_num();
    // printf("num_layers = %i\n", num_layers);
    //Find largest layer && ensure all layers initialized
    int max = 0;
    for(int i = 0; i < num_layers; i++){
      if(max < model->get_layer(i)->neurons){
        max = model->get_layer(i)->neurons;
      }
    }
    //
    // Fire parent kernel
    // 
    float* inputs;
    gpuErrchk( cudaMalloc((void**)&inputs, model->get_layer(0)->neurons*sizeof(float)));
    gpuErrchk( cudaMemcpy(inputs,in,model->get_layer(0)->neurons*sizeof(float),cudaMemcpyHostToDevice));   
    //
    // printf("\n-------EXEC------\n");
    //
    float* output_buffer_even;
    float* output_buffer_odd;
    //
    gpuErrchk( cudaMalloc((void**)&output_buffer_even, max*sizeof(float)) ); // Test if it'd be better to init a sized buffer for each layer or stick with a max
    gpuErrchk( cudaMalloc((void**)&output_buffer_odd, max*sizeof(float)) ); // Test if it'd be better to init a sized buffer for each layer or stick with a max
    //
    // int start = time_stamp();
    //
    for(int l = 0; l < num_layers; l++){
      //
      // printf("\n- layer %i ----------\n\n", l);
      LayerDefinition* layer = model->get_layer(l);
      //
      //
      int N = layer->neurons;
      int prevN = layer->prev_neurons;
      //
      //
      bool even = ieven(l);
      // 
      if((N*prevN)<=1024){
        //
        //Exec in one pass
        //  
        dim3 threadsPerBlock(N, prevN);
        dim3 numBlocks(1, 1);
        //
        //
        if(l == 0){        //ROOT case 
          //
          fast_layer_forward<<<numBlocks, threadsPerBlock>>>(N, inputs, output_buffer_even, layer->cuda_weights, layer->cuda_biases);
          gpuErrchk( cudaPeekAtLastError() );
          gpuErrchk( cudaDeviceSynchronize() );
          gpuErrchk( cudaFree(inputs) );
          //
        }else if(even){    //EVEN case
          //
          fast_layer_forward<<<numBlocks, threadsPerBlock>>>(N, output_buffer_odd, output_buffer_even, layer->cuda_weights, layer->cuda_biases);
          gpuErrchk( cudaPeekAtLastError() );
          gpuErrchk( cudaDeviceSynchronize() );
          //
          }else if(!even){ //ODD  case
          //
          fast_layer_forward<<<numBlocks, threadsPerBlock>>>(N, output_buffer_even, output_buffer_odd, layer->cuda_weights, layer->cuda_biases);
          gpuErrchk( cudaPeekAtLastError() );
          gpuErrchk( cudaDeviceSynchronize() );
          //
        } 
        
      }else{
        //
        //Exec in chunks
        int gridY = 0; 
        int blockX = 0;
        int blockY = 0;
        //
        //
        int n_d = N; 
        int p_d = prevN;
        // printf("N: %i, P: %i\n", N, prevN);
        //
        //
        if(prevN>N||prevN==N){
          //
          n_d = N;
          p_d = prevN/N;
          reduce_fraction(p_d, n_d);
          //
        }else{
          //
          n_d = N;
          p_d = N/prevN;
          reduce_fraction(p_d, n_d);
          //
        }          
        //
          gridY = (prevN/max_threads)+1;
        if(prevN<=max_threads){
          //gridY++;
          blockX = (prevN);
          blockY = 1;
          //
        }else{
          //
          blockX = max_threads/(p_d);
          blockY = (max_threads)/(max_threads/(p_d));
          //
        }
        //
        // printf("n: %i p: %i \nblocksPerGrid(%i, %i)\nthreadsPerBlock(%i, %i)\n", n_d, p_d, N, gridY, blockX, blockY);
        //  
        // blocks per neuron = prevN / 1024
        //
        dim3 numBlocks(N, gridY);
        dim3 threadsPerBlock(blockX, blockY);
        //10240 inputs = 10 blocks per neuron
        // 

        //
        //
        // cudaError_t err;
        if(l == 0){     //ROOT case 
          fast_4d_forward<<<numBlocks, threadsPerBlock>>>(prevN, inputs, output_buffer_even, layer->cuda_weights, layer->cuda_biases);
          gpuErrchk( cudaPeekAtLastError() );
          gpuErrchk( cudaDeviceSynchronize() );
          gpuErrchk( cudaFree(inputs) );
        }else if(even){ //EVEN case
          fast_4d_forward<<<numBlocks, threadsPerBlock>>>(prevN, output_buffer_odd, output_buffer_even, layer->cuda_weights, layer->cuda_biases);
          gpuErrchk( cudaPeekAtLastError() );
          gpuErrchk( cudaDeviceSynchronize() );
        }else{          //ODD  case
          fast_4d_forward<<<numBlocks, threadsPerBlock>>>(prevN, output_buffer_even, output_buffer_odd, layer->cuda_weights, layer->cuda_biases);
          gpuErrchk( cudaPeekAtLastError() );
          gpuErrchk( cudaDeviceSynchronize() );
        } 
      }
      //
      // copy last layer output
      if(l == num_layers-1){
        if(even){ 
          //EVEN case
          gpuErrchk(  cudaMemcpy(out,output_buffer_even,layer->neurons*sizeof(float),cudaMemcpyDeviceToHost) );
        }else{    
          //ODD  case
          gpuErrchk(  cudaMemcpy(out,output_buffer_odd,layer->neurons*sizeof(float),cudaMemcpyDeviceToHost) );
          
        } 
      }
      //
      // 
      //

    } // End layers
    gpuErrchk( cudaFree(output_buffer_odd) );
    gpuErrchk( cudaFree(output_buffer_even) );
    gpuErrchk( cudaPeekAtLastError() );
    gpuErrchk( cudaDeviceSynchronize() );
    //
    return;
  }




  //Model permutation methods

  char* newModel(int ModelLibraryID, int LayerNum){
    //Create a new model
    ModelDefinition* model = new(ModelDefinition)(ModelLibraryID,LayerNum);
    model->init_layers();
    return (char*)model;

  }

  int appendLayer(char* ptr, const int N, const int prevN, float* weights, float* biases){
    //   
    // Cast ptr to model
    ModelDefinition* modelDefPtr = (ModelDefinition*)ptr;
    //
    LayerDefinition layerDef = LayerDefinition();
    layerDef.layer_index = modelDefPtr->get_append_index();
    layerDef.neurons = N; 
    layerDef.ready = false;
    layerDef.prev_neurons = prevN; 
    // //
    //GPU alloc sizes
    size_t weight_size = (N*prevN)* sizeof(float);
    size_t biases_size = N * sizeof(float);
    //
    //Alloc on GPU
    gpuErrchk(cudaMalloc((void**)&layerDef.cuda_weights, weight_size));
    gpuErrchk(cudaMemcpy(layerDef.cuda_weights,weights,weight_size,cudaMemcpyHostToDevice));
    //
    gpuErrchk(cudaMalloc((void**)&layerDef.cuda_biases, biases_size));
    gpuErrchk(cudaMemcpy(layerDef.cuda_biases,biases,biases_size,cudaMemcpyHostToDevice));
    //
    //
    modelDefPtr->append_layer(layerDef);
    return 0;
  }



  // Model data methods

  int getModelID(char* ptr){
    //Append new layer to previous layer of model
    ModelDefinition* modelPtr = (ModelDefinition*)ptr;
    return (int)modelPtr->get_id();
  }


  void disposeModel(char* ptr){
    //
    ModelDefinition* modelPtr = (ModelDefinition*)ptr;
    modelPtr->dispose();
    return;
  }

  // Device Methods

  void printCudaInfo(void){
    int nDevices;
    // Use this code later ;) 
    cudaGetDeviceCount(&nDevices);
    //
    for (int i = 0; i < nDevices; i++) {
      cudaDeviceProp prop;
      
      cudaGetDeviceProperties(&prop, i);
      printf("Device Number: %d\n", i);
      printf("  Max Threads: %i\n", prop.maxThreadsPerBlock);
      printf("  Memory Clock Rate (KHz): %d\n",
            prop.memoryClockRate);
      printf("  Memory Bus Width (bits): %d\n",
            prop.memoryBusWidth);
      printf("  Peak Memory Bandwidth (GB/s): %f\n\n",
            2.0*prop.memoryClockRate*(prop.memoryBusWidth/8)/1.0e6);
    }
  }
}

