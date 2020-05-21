Pod::Spec.new do |spec|
  spec.name         = 'mfachain'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/MFAChain/mfachain'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS MFA Client'
  spec.source       = { :git => 'https://github.com/MFAChain/mfachain.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/mfachain.framework'

	spec.prepare_command = <<-CMD
    curl https://mfastore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/mfachain.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
