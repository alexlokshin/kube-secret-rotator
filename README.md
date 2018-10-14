# kube-secret-rotator

Do you have a need to rotate secret values? Do you sign something with a secret only to have the same exact secret in production for years at a time? Does security breath down your neck with no funds to implement a proper secret vault? `kube-secret-rotator` is an answer.

Given the desired frequency of rotation, it can rotate one or more secrets, saving off the old value into a separate key. This way you can validate past credentials even if the secret value has rolled. Secrets are created automatically if not present.

Each secret's rotation strategy is specified by the following parameters:

* NAME - secret name
* NAMESPACE - namespace to rotate the secret in
* KEY - a key within a kubernetes secret that will be modified
* STRATEGY - new value creation strategy 