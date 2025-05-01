'''Test Docker image tools.'''

from test.util_docker import clean_docker

from manage.docker import image


def test_image_build(request, tmp_file, uncap):
    dockerfile = tmp_file('Dockerfile',
                          'FROM alpine:latest',

                          # Test build args pass through
                          'ARG arg',
                          'RUN test "$arg" = "value"',

                          # Test files adjacent to the Dockerfile are part of the build context
                          'COPY some-file some-file',
                          'RUN test -f some-file')

    tmp_file('some-file', 'Hello!')

    name = request.node.name

    build_args = {
        'arg': 'value'
    }

    image_, log = image.build(dockerfile, name, build_args)

    uncap(log)

    assert image_ is not None

    clean_docker(name)
