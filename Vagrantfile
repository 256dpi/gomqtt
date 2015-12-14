Vagrant.configure(2) do |config|
  config.vm.box = 'ubuntu/trusty64'
  config.vm.network 'forwarded_port', guest: 1884, host: 1885
  config.vm.network 'private_network', ip: '10.0.3.15'
end
