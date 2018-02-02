# RT vs Capistrano, Fabric etc.

Capistrano comes from the Ruby community and Fabric from the Python community.

Both tools allow easy deployment for applications written in that language.
It might be possible to define any deployment we might be able to define via RT/Terraform,
however the long-term maintainability and reusability of such solution may differ.

Terraform's modular architecture that we can use in RT allows us to define templates that
can be easily reused across teams. These templates are written in HCL which (intentionally)
limits what user can do in the language.

It may be possible to define more advanced language constructs like loops and conditionals
in Fabric or Capistrano, whether such constructs make such code readable, maintainable
and easily reusable is another story.

---------

_TODO: This needs more care from someone who knows more about Capistrano or Fabric_
