# Summary of achievement for the type_checking proposal

Checking and adapting the vula python code to fully support type-checking with `mypy`, including `--strict` have been fulfilled. Use `make mypy` within the `/podman` directory of this project. the stubs modules have been added to the `Pipfile` within the root directory of this project and is installed within the images when using the previous command utilizing podman. 
The type `Any` has only rarely been used for making the code comply with mypy.

Enforcing type checking ... (open minimum goal) TODO

Additional type-check pytest for cross-validation has been implemented using the `make pyright` executed from the `/podman` directory of this project. 


## Summary of Goals


| status | type     | description                                                                                                                              |
|--------| -------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| ok     | minimum  | Resolve all issues raised by `mypy –strict`                                                                                              |
| ok     | minimum  | Add a second type checker for cross verification: [pyright]([[vula type_checking.pdf#page=1&annotation=8R\|vula type_checking, page 1]]) |
| ok     | minimum  | Add the type checkers to the CI configurations by extend the existing command ‘make check‘.                                              |
| TODO   | minimum  | Add a git pre-commit hook which runs the type checkers on each commit (e.g.: [pre-commit](https://pre-commit.com/)).                     |
| ok     | optioal  | Add python type stubs for dependencies (Ad-hoc type stubs in the vula codebase)                                                          |
| ?      | optional | Add type annotations for dependencies (Contribution to a third party library)                                                            |


Future work may include to adapt the typing to fully comply with pytest type standards.