

# kubectl get pod --all-namespaces --sort-by='.metadata.name' -o json | jq -r '[.items[] | {ns: .metadata.namespace, containers: .spec.containers[] | [ {container_name: .name, memory_requested: .resources.requests.memory,memory_limits: .resources.limits.memory, cpu_requested: .resources.requests.cpu,cpu_limits: .resources.limits.cpu } ] }]' | jq  'sort_by(.containers[0].cpu_requested)' | jq  'sort_by(.ns)' |
kubectl get pod --all-namespaces --sort-by='.metadata.name' -o json | jq -r '[.items[] | {name:.metadata.name, ns: .metadata.namespace,state: .status.phase, containers: .spec.containers[] | [ {memory_requested: .resources.requests.memory,memory_limits: .resources.limits.memory, cpu_requested: .resources.requests.cpu,cpu_limits: .resources.limits.cpu } ] }]' | jq  'sort_by(.containers[0].cpu_requested)' | jq  'sort_by(.ns)' | \
python3 -c "
import json,sys
from sympy.parsing.sympy_parser import parse_expr,implicit_multiplication_application,standard_transformations,factorial_notation
transformations = (standard_transformations + (implicit_multiplication_application,)+(factorial_notation,))

a=json.loads(sys.stdin.read())
memory_requested='0'
memory_limits='0'
cpu_requested='0'
cpu_limits='0'
for i in a:
    for ik,iv in i['containers'][0].items():
        if ik == 'memory_requested' and iv:
#            print(str(iv) +'+'+ str(memory_requested))
            memory_requested=parse_expr(str(iv) +'+'+ str(memory_requested),transformations=transformations)
        if ik == 'memory_limits' and iv:
#            print(str(iv) +'+'+ str(memory_limits))
            memory_limits=parse_expr(str(iv) +'+'+ str(memory_limits),transformations=transformations)
        if ik == 'cpu_requested' and iv:
#            print(str(iv) +'+'+ str(cpu_requested))
            cpu_requested=parse_expr(str(iv) +'+'+ str(cpu_requested),transformations=transformations)
        if ik == 'cpu_limits' and iv:
#            print(str(iv) +'+'+ str(cpu_limits))
            cpu_limits=parse_expr(str(iv) +'+'+ str(cpu_limits),transformations=transformations)
    print('\n')
#print(memory_requested.subs('i','1000/1024').evalf(),memory_limits.subs('i',1).subs('M','G*(1/1024)').evalf(),cpu_requested,cpu_limits,sep='\n')
print(memory_requested.subs('i','1000/1024').evalf(),memory_limits.subs('i',1).subs('M','G*(1/1024)').evalf(),cpu_requested.subs('m','0.001').evalf(),cpu_limits.subs('m','0.001').evalf(),sep='\n')
"