This upgrade upgrades ArgoCD by
* deleting the Helm Release 
* deleting all secrets
* Re-creating the above, with the up-to-date create method from okctl.

All secrets are deleted and re-created because they have a new format.
