# kubeutil
useful utils for kubernetes related development.    

1. a controller demo for watch pod event and execute handle function.     
   it works just like buildin controllers in kubernetes and the code too.    

2. implemented webshell to pod in kubernetes cluster.     
   webshell url example: http://127.0.0.1:8090/terminal?namespace=default&pod=nginx-65f9798fbf-jdrgl&container_name=nginx    
   [introduction](http://maoqide.live/post/cloud/kubernetes-webshell/)    

# plan    
TODO:    
- [x] using go mod    
- [ ] refactor modules    
- [ ] operator template codes    
- [ ] refactor kubeboxs    
- [ ] more...    