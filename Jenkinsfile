pipeline {
    agent any

    stages {
        stage ('Pre-Build') {
            steps {
				sh 'go get -tags tools || true'
            }
        }
        
        stage('Compile') {
            steps {
				withEnv(["PATH+GO=${root}/go/bin"]) {
					sh 'go env'
					sh 'echo $PATH'
                	sh 'buf generate'
	                sh 'make daemon-compile'
				}
            }
        }
        
        stage ('Post-Compile') {
            parallel { 
                stage('Codestyle') {
                    steps {
                        sh 'make daemon-codestyle'
                        sh 'make webui-codestyle'
                    }
                }
                stage('UnitTests') {
                    steps {
                        sh 'make daemon-unittests'
                    }
                }
            }
        }
        
    }
}
