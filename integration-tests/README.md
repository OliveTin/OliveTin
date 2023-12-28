# OliveTin-integration-tests

## GitHub Actions (Ubuntu, Local Process)

- `mocha` is run with the default runner that starts and stops OliveTin as a local process (ie, localhost:1337).

## Running different configurations (Local Process, VM, Container)

- Get the snapshot you want to test `make getsnapshot`
- To test against VMs:
-- `export OLIVETIN_TEST_RUNNER=container`
-- `vagrant up f38` (or whatever distro you like defined in `Vagrantfile`)
-- `. envVagrant.sh f38` to set the $IP and $PORT
- `mocha`
