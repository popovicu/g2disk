"""Generates the version header file using a template.

In typical nbdkit build, this is done with autoconf.
"""

def _expand_version_impl(ctx):
    output = ctx.outputs.expanded_file

    ctx.actions.expand_template(
        template = ctx.file.template,
        output = output,
        substitutions = {
            "@NBDKIT_VERSION_MAJOR@": ctx.attr.major,
            "@NBDKIT_VERSION_MINOR@": ctx.attr.minor,
            "@NBDKIT_VERSION_MICRO@": ctx.attr.micro,
        },
    )

    return [DefaultInfo(files = depset([output]))]

expand_version = rule(
    implementation = _expand_version_impl,
    attrs = {
        "major": attr.string(mandatory = True),
        "minor": attr.string(mandatory = True),
        "micro": attr.string(mandatory = True),
        "template": attr.label(
            mandatory = True,
            allow_single_file = True,
        ),
        "expanded_file": attr.output(mandatory = True),
    },
)