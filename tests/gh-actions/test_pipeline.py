import kfp
from kfp import dsl
import kfp.components as comp


@comp.create_component_from_func
def echo_op():
    print("Test pipeline")

@dsl.pipeline(
    name='test-pipeline',
    description='A test pipeline.'
)
def hello_world_pipeline():
    echo_task = echo_op()

if __name__ == '__main__':
    kfp.compiler.Compiler().compile(hello_world_pipeline, __file__ + '.yaml')
