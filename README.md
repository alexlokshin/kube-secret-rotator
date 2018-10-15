# kube-secret-rotator

Do you have a need to rotate secret values? Do you sign something with a secret only to have the same exact secret in production for years at a time? Does security breath down your neck with no funds to implement a proper secret vault? `kube-secret-rotator` is an answer.

Given the desired frequency of rotation, it can rotate one or more secrets, saving off the old value into a separate key. This way you can validate past credentials even if the secret value has rolled. Secrets are created automatically if not present.

When running on Kubernetes, `kube-secret-rotator` expects `-secret` and `-frequency` parameters, like this:

`kube-secret-rotator -secret=NAME,NAMESPACE,KEY,STRATEGY -frequency=60`

In case multiple values are rotated, you can specify it like this:

`kube-secret-rotator -secret=NAME,NAMESPACE,KEY,STRATEGY|NAME,NAMESPACE,KEY,STRATEGY|...|NAME,NAMESPACE,KEY,STRATEGY -frequency=60`

Frequency here is specified in minutes.

Each secret's rotation strategy is specified by the following parameters:

* `NAME` - secret name
* `NAMESPACE` - namespace to rotate the secret in
* `KEY` - a key within a kubernetes secret that will be modified
* `STRATEGY` - new value creation strategy (supported values: `retainPrev`, `omitPrev`). `retainPrev` will save off an old value in the key within the same secret, appending `_PREV` to the name of the key. If `omitPrev` is specified, the old value is not saved.

In case you need to rotate different secrets at different schedules, deploy several instances of `kube-secret-rotator`.