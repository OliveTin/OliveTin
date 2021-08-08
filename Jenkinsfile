pipeline {
    agent any

    stages {
        stage ('Pre-Build') {
            steps {
                sh 'make grpc'
            }
        }
        
        stage('Compile') {
            steps {
                sh 'make grpc daemon-compile'
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
