# Caution and Warning

!!! warning
    **We strongly advise you to not run this application on any Azure tenant, where you cannot afford to lose
    all resources.**

To reduce the blast radius of accidents, there are some safety precautions:

1. By default, **azure-nuke** only lists all nuke-able resources. You need to add `--no-dry-run` to actually delete
   resources.
2. **azure-nuke** asks you twice to confirm the deletion by entering the account alias. The first time is directly
   after the start and the second time after listing all nuke-able resources.
       
    !!! note "ProTip"
        This can be disabled by adding `--no-prompt` to the command line. 

3. The config file contains a blocklist field. If the Account ID of the account you want to nuke is part of this
   blocklist, **azure-nuke** will abort. It is recommended, that you add every production account to this blocklist.
4. To ensure you don't just ignore the blocklisting feature, the blocklist must contain at least one Account ID.
5. The config file contains account specific settings (e.g. filters). The account you want to nuke must be explicitly
   listed there.
6. To ensure to not accidentally delete a random tenant, it is required to specify a config file. It is recommended
   to have only a single config file and add it to a central repository. This way the blocklist is easier to manage and
   keep up to date.

Feel free to create an issue, if you have any ideas to improve the safety procedures.

