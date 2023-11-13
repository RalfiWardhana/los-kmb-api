// ::VARIABLE DEFENITION
def serviceName             = "los-kmb-api"
def fullname                = "${JOB_BASE_NAME}-${BUILD_NUMBER}"

// ::GIT URL
def codeUrl                 = "https://github.com/KB-FMF/los-kmb-api.git"
def configUrl_devtest       = "https://github.com/KB-FMF/los-devops.git"
def configUrl_stgprd        = "https://github.com/KB-FMF/devops-config.git"

// ::GO DEPENDENCIES
def go_path                 ="/go"
def go_root                 ="/usr/local/go"
def path                    ="/go/bin:/usr/local/go/bin"
def urlGo                   = "https://storage.googleapis.com/golang/go1.17.1.linux-amd64.tar.gz"

// ::CONFIG FOLDER
def devFolder               = "config-development/config-los-kmb-api"
def testFolder              = "config-testing/config-los-kmb-api"
def stgFolder               = "config-staging/config-los-kmb-api"
def prdFolder               = "prd/los/los-kmb-api"

// ::CREDENTIALS
def garCred                 = "sa-gar"
def gchatKey                = "gchat-api-key"
def gchatToken              = "gchat-token-los"
def gitCred                 = "github-cred"
def nrCred                  = "nr-api-key"
def sonarqubeCred           = "sonarqube-token"
def vmCred_dev              = "dev-los-cred"
def vmCred_test             = "test-los-cred"
def vmCred_stg              = "stg-los-cred"
def vmCred_prd              = "prd-los-cred"

// ::DEPLOYMENT MARKER
def appId_dev               = ""
def appId_test              = ""
def appId_stg               = ""
def appId_prd               = ""

// ::NOTIFICATION
def devstgChannel           = "los-dev-builds"
def prdSlack                = "los-builds"
def gchatRoomID             = "gchat-room-id-los"
def gchatUrl                = "gchat-url-los"
def gchatWorkspace          = "https://chat.googleapis.com/v1/spaces/${gchatRoomID}/messages?key=${gchatKey}&token=${gchatToken}, id:${gchatUrl}"

// ::INITIALIZATION
def getCommit_an, getCommit_id, unitTest_score

podTemplate(
    label: fullname,
    containers: [
        //containerTemplate(name: 'curl', image: 'centos', command: 'cat', ttyEnabled: true),
        containerTemplate(name: 'golang', image: 'golang:1.17', command: 'cat', ttyEnabled: true),
        containerTemplate(name: 'ubuntu', image: 'ubuntu:16.04',command: 'cat',ttyEnabled: true)
    ]
) 

{
    node(fullname) {
        try{
            // ::TRIGGER && VERSIONING
            if (env.ENVIRONMENT == "dev") {
                environment     = "Development"
                serviceVersion  = "${Tag}"
                sourceCode      = "${Tag}"
                appId           = "${appId_dev}"
                configFolder    = "${devFolder}"
                configUrl       = "${configUrl_devtest}"
                targetVM        = "${devVM}"
                credVm          = "${vmCred_dev}"
                portVm          = "${devPort}"
                projectName     = "${env.ENVIRONMENT}-${serviceName}:${serviceVersion}"
                projectKey      = "${env.ENVIRONMENT}-${serviceName}"
                refs            = "+refs/tags/*:refs/remotes/origin/tags/*"
                slackChannel    = "${devstgChannel}"
                echo "Running ${environment} pipeline with tag version: ${serviceVersion}"
            } else if (env.ENVIRONMENT == "testing") {
                environment     = "Testing"
                serviceVersion  = "${Tag}"
                sourceCode      = "${Tag}"
                appId           = "${appId_test}"
                configFolder    = "${testFolder}"
                configUrl       = "${configUrl_devtest}"
                targetVM        = "${testVM}"
                credVm          = "${vmCred_test}"
                portVm          = "${testPort}"
                projectName     = "${env.ENVIRONMENT}-${serviceName}:${serviceVersion}"
                projectKey      = "${env.ENVIRONMENT}-${serviceName}"
                refs            = "+refs/tags/*:refs/remotes/origin/tags/*"
                slackChannel    = "${devstgChannel}"
                echo "Running ${environment} pipeline with tag version: ${serviceVersion}"
            } else if (env.ENVIRONMENT == "stg") {
                environment     = "Staging"
                serviceVersion  = "${Tag}"
                sourceCode      = "${Tag}"
                appId           = "${appId_stg}"
                configFolder    = "${stgFolder}"
                configUrl       = "${configUrl_stgprd}"
                targetVM        = "${stgVM}"
                credVm          = "${vmCred_stg}"
                portVm          = "${stgPort}"
                projectName     = "${env.ENVIRONMENT}-${serviceName}:${serviceVersion}"
                projectKey      = "${env.ENVIRONMENT}-${serviceName}"
                refs            = "+refs/tags/*:refs/remotes/origin/tags/*"
                slackChannel    = "${devstgChannel}"
                echo "Running ${environment} pipeline with tag version: ${serviceVersion}"
            } else if (env.ENVIRONMENT == "prd") {
                environment     = "Production"
                serviceVersion  = "${Tag}"
                sourceCode      = "${Tag}"
                appId           = "${appId_prd}"
                configFolder    = "${prdFolder}"
                configUrl       = "${configUrl_stgprd}"
                targetVM1       = "${prdVM1}"
                targetVM2       = "${prdVM2}"
                credVm          = "${vmCred_prd}"
                portVm          = "${prdPort}"
                projectName     = "${env.ENVIRONMENT}-${serviceName}:${serviceVersion}"
                projectKey      = "${env.ENVIRONMENT}-${serviceName}"
                refs            = "+refs/tags/*:refs/remotes/origin/tags/*"
                slackChannel    = "${prdSlack}"
                echo "Running ${environment} pipeline with tag version: ${serviceVersion}"
            } else {
                environment     = "Pull Request"
                serviceVersion  = "pr-${ghprbPullId}"
                sourceCode      = "**/pr/${ghprbPullId}/*"
                projectName     = "pr-${serviceName}:${serviceVersion}"
                projectKey      = "pr-${serviceName}"
                refs            = "+refs/pull/*:refs/remotes/origin/pr/*"
                slackChannel    = "${devstgChannel}"
               echo "Running ${environment} pipeline with PR Number: ${ghprbPullId}"
            }

            currentBuild.displayName="#${BUILD_NUMBER} ${JOB_BASE_NAME}: ${serviceVersion}"
            sendNotification(job_status: "START", name: serviceName, version: serviceVersion, to: environment, gchat: gchatWorkspace, channel: slackChannel)
            
            stage('Checkout') {
                stage = 'Checkout'
                echo "Create directory for source code and config"
                sh "mkdir code config"
                parallel(
                    'Code': {
                        dir('code') {
                            // ::SOURCE CODE CHECKOUT
                            echo "Checkout ${serviceName} source code"
                            def scm = checkout([
                                $class: 'GitSCM',
                                branches: [[name: sourceCode]],
                                userRemoteConfigs: [[credentialsId: gitCred, url: codeUrl, refspec: refs]]
                            ])
                        }
                    },

                    'Config': {
                        dir('config') {
                            if (environment != "Pull Request") {
                                // ::CONFIG FILE CHECKOUT
                                echo "Checkout ${serviceName} config files"
                                def scm = checkout([
                                    $class: 'GitSCM',
                                    branches: [[name: 'master']],
                                    userRemoteConfigs: [[credentialsId: gitCred, url: configUrl]]
                                ])
                            } else {
                                echo "Pull Request doesn't checkout config files"
                            }
                        }
                    }
                )
            }

            // ::DEV & STG
            if (environment != "Production") {
                dir('code') {
                    /*stage('Unit Test') {
                        stage = 'Unit Test'
                            container('golang') {
                            echo "Running unit test"
                            gitConfig(credentials: gitCred)
                            goEnv()
                            unitTest(name: serviceName, version: serviceVersion)
                            try {
                                unitTest_publish()
                            } catch(e) {
                                echo "Unit test doesn't exist"
                                currentBuild.result = "UNSTABLE"
                            }
                            unitTest_score = unitTest_result()
                            unitTestGetValue = "Your score is ${unitTest_score}"
                            echo "${unitTestGetValue}"
                            if (unitTest_score >= env.UNITTEST_STANDARD) {
                                echo "Unit test fulfill standar value with score ${unitTest_score}/${env.UNITTEST_STANDARD}"
                            } else {
                                currentBuild.result = "ABORTED"
                                error("Sorry your unit test score not fulfill standard with score ${unitTest_score}/${env.UNITTEST_STANDARD}")
                            }
                        }
                    }*/

                    stage('Code Review') {
                        stage = 'Code Review'
                        echo "Running Code Review with SonarQube"
                        def scannerHome = tool "sonarscanner"
                        withSonarQubeEnv (credentialsId: sonarqubeCred, installationName: "SonarQube") {
                            sonarScan(name: serviceName, scannerHome: scannerHome, project_name: projectName, project_key: projectKey, project_version: serviceVersion)
                        }
                        timeout(time: 10, unit: 'MINUTES') {
                            waitForQualityGate abortPipeline: true
                        }
                    }
                }
            }
        
            if (environment != "Pull Request") {
                container('ubuntu') {
                    stage('Install Dependencies') {
                        stage = 'Install Depedencies'
                        echo 'Install all dependencies service ${serviceName} version: ${serviceVersion}'
                        initiate()
                        sshSetup()
                        goSetup(goUrl: urlGo, pathgo: go_path , rootgo: go_root, paths: path)
                    }
                    
                    stage('Build App') {
                        stage = 'Build App'
                        withEnv(["GOPATH=${go_path}","GOROOT=${go_root}","PATH+GO=${path}"]){
                            stage = 'Build App'
                            echo 'Build service ${serviceName} version: ${serviceVersion} to binary'
                            gitConfig(credentials: gitCred)
                            buildApp(name: serviceName, configPath: configFolder, versionApp: serviceVersion)
                        }
                    }
                    
                    stage('Deployment') {
                        stage = 'Deployment'
                        echo "Deploy service ${serviceName} version: ${serviceVersion} to ${environment} environment"
                        if (environment != "Production") {
                            deployApp(vmCred: credVm, instanceIP: targetVM, name: serviceName, port: portVm)
                        } else {
                            deployApp(vmCred: credVm, instanceIP1: targetVM1, instanceIP2: targetVM2, name: serviceName, port: portVm)
                        }
                    }
                
                    stage('Restart Service') {
                        stage = 'Restart Service'
                        if (environment != "Production") {
                            restartService(vmCred: credVm, instanceIP: targetVM, name: serviceName, port: portVm)
                        } else {
                            restartService(vmCred: credVm, instanceIP1: targetVM1, instanceIP2 :targetVM2, name: serviceName, port: portVm)
                        }
                    }
                }
                //container('curl') {
                    //deploymentMarker(credentials: nrCred, id: appId, version: serviceVersion)
                //}
            } else {
                echo "No deployment for Pull Request"
            }

            stage('Notification') {
                dir('code') {
                    echo "Job Success"
                    getCommit_an    = gitCommit_authorName()
                    getCommit_id    = gitCommit_id()
                    sendNotification(job_status: currentBuild.currentResult, name: serviceName, version: serviceVersion, gitcommit_an: getCommit_an, gitcommit_id: getCommit_id, unittest_score: unitTest_score, unittest_standard: env.UNITTEST_STANDARD, to: environment, gchat: gchatWorkspace, channel: slackChannel)
                }
            }
        } catch(e) {
            currentBuild.result = "FAILURE"
            echo "${e}"
            echo "Job Fail"
            sendNotification(job_status: "FAILURE", name: serviceName, version: serviceVersion, gitcommit_an: getCommit_an, gitcommit_id: getCommit_id, unittest_score: unitTest_score, unittest_standard: env.UNITTEST_STANDARD, to: environment, gchat: gchatWorkspace, channel: slackChannel, error: "${e}", stage: "${stage}")
        }
    }
}

// ::FUNCTION
def buildApp(Map args) {
    sh "mkdir /opt/go/ /opt/go/src/ /opt/go/src/${args.name}/ /opt/go/src/${args.name}/conf/"
    sh "cp ./config/${args.configPath}/${env.ENVIRONMENT}.env /opt/go/src/${args.name}/conf/config.env"
    sh "sed -i 's/tagVersion/'${args.versionApp}'/g' /opt/go/src/${args.name}/conf/config.env"
    sh "cp -a ./code/* /opt/go/src/${args.name}/"
    sh "cd /opt/go/src/${args.name}/ && go mod vendor"
    sh "cd /opt/go/src/${args.name}/app/ && go build -o /opt/go/src/${args.name}/${args.name}"
    sh "cd /opt/go/src/ && tar -czvf ${args.name}.tar.gz ./${args.name} "
}

def deployApp(Map args) {
    withCredentials([usernamePassword(credentialsId: "${args.vmCred}",passwordVariable:'PassCred',usernameVariable:'UserCred')]) {
        if (environment != "Production") {
            echo "Deploy to ${args.instanceIP}"
            sh """
                sshpass -p '${PassCred}' ssh -o "StrictHostKeyChecking=no" -p ${args.port} ${UserCred}@${args.instanceIP} "hostname"
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP} "hostname"
                sshpass -p '${PassCred}' scp -P ${args.port} /opt/go/src/${args.name}.tar.gz ${UserCred}@${args.instanceIP}:/opt/temp/
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP} "sudo tar -xzf /opt/temp/${args.name}.tar.gz -C /opt/go/src/ && rm /opt/temp/${args.name}.tar.gz"
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP} "sudo chown -R los-admin:los-admin /opt/go/src/${args.name}/"
            """
        } else {
            echo "Deploy to ${args.instanceIP1}"
            sh """
                sshpass -p '${PassCred}' ssh -o "StrictHostKeyChecking=no" -p ${args.port} ${UserCred}@${args.instanceIP1} "hostname"
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP1} "hostname"
                sshpass -p '${PassCred}' scp -P ${args.port} /opt/go/src/${args.name}.tar.gz ${UserCred}@${args.instanceIP1}:/opt/temp/
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP1} "sudo tar -xzf /opt/temp/${args.name}.tar.gz -C /opt/go/src/ && rm /opt/temp/${args.name}.tar.gz"
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP1} "sudo chown -R los-admin:los-admin /opt/go/src/${args.name}/"
            """
            echo "Deploy to ${args.instanceIP2}"
            sh """
                sshpass -p '${PassCred}' ssh -o "StrictHostKeyChecking=no" -p ${args.port} ${UserCred}@${args.instanceIP2} "hostname"
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP2} "hostname"
                sshpass -p '${PassCred}' scp -P ${args.port} /opt/go/src/${args.name}.tar.gz ${UserCred}@${args.instanceIP2}:/opt/temp/
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP2} "sudo tar -xzf /opt/temp/${args.name}.tar.gz -C /opt/go/src/ && rm /opt/temp/${args.name}.tar.gz"
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP2} "sudo chown -R los-admin:los-admin /opt/go/src/${args.name}/"
            """
        }
    }
}

def deploymentMarker(Map args) {
    withCredentials([string(credentialsId: "${args.credentials}", variable: 'ApiKey')]) {
        sh """
            #!/bin/bash
            curl -X POST "https://api.newrelic.com/v2/applications/${args.id}/deployments.json" \
            -H "X-Api-Key:${ApiKey}" -i \
            -H "Content-Type: application/json" \
            -d \
            '{
                "deployment": {
                    "revision": "${args.version}",
                    "changelog": "",
                    "description": "",
                    "user": "Automated Sent By Jenkins"
                }
            }' 
        """
    }
}

def gitCommit_id() {
    sh(script: 'git rev-parse HEAD', returnStdout: true).trim()
}

def gitCommit_authorName() {
    sh(script: 'git log -1 --pretty=format:"%an"', returnStdout: true).trim()
}

def gitConfig(Map args) {
    withCredentials([usernamePassword(credentialsId: "${args.credentials}", passwordVariable: 'password', usernameVariable: 'username')]) {
        sh 'git config --global url."https://${username}:${password}@github.com".insteadOf "https://github.com"'
    }
}

def goEnv() {
   // sh "go env -w GO111MODULE=on"
   // sh "go env -w CGO_ENABLED=1"
   // sh "go env -w GOOS=linux"
    sh "go install github.com/jstemmer/go-junit-report/v2@latest"
}

def goSetup(Map args) {
    sh "curl -s ${args.goUrl}| tar -v -C /usr/local -xz > /dev/null 2>&1"
    sh "export GOPATH=${args.pathgo}"
    sh "export GOROOT=${args.rootgo}"
    sh "export PATH=$PATH:${args.paths}"
    sh "export PATH=$PATH:${args.rootgo}/bin"
    sh "export PATH=$PATH:${args.pathgo}/bin"
}

def initiate() {
    sh "apt-get update -qy > /dev/null 2>&1"
    sh "apt-get upgrade -y -q > /dev/null 2>&1"
    sh "apt-get install -y -q curl build-essential ca-certificates git > /dev/null 2>&1"
}

def restartService(Map args) {
    withCredentials([usernamePassword(credentialsId: "${args.vmCred}", passwordVariable: 'PassCred', usernameVariable: 'UserCred')]) {
        if (environment != "Production") {
            sh """
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP} "sudo systemctl stop ${args.name}"
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP} "sudo systemctl start ${args.name}"
            """
        } else {
            sh """
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP1} "sudo systemctl stop ${args.name}"
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP1} "sudo systemctl start ${args.name}"
            """
            sh """
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP2} "sudo systemctl stop ${args.name}"
                sshpass -p '${PassCred}' ssh -p ${args.port} ${UserCred}@${args.instanceIP2} "sudo systemctl start ${args.name}"
            """
        }
    }	
}

def sendNotification(Map args) {
    // Messages
    message_start = """*${args.job_status}* CICD Pipeline with details :
```Job             : ${env.JOB_NAME}
Service Name    : ${args.name}
Environment     : ${args.to}
Build Number    : ${env.BUILD_NUMBER}
Service Version : ${args.version}```
More info at    : ${env.BUILD_URL}
"""
    if (environment != "Production") {
        message_success = """*${args.job_status}* CICD Pipeline with details :
```Job             : ${env.JOB_NAME}
Service Name    : ${args.name}
Environment     : ${args.to}
Build Number    : ${env.BUILD_NUMBER}
Service Version : ${args.version}
Commit Author   : ${args.gitcommit_an}
Commit ID       : ${args.gitcommit_id}
Unit Test Result: ${args.job_status} with score ${args.unittest_score}/${args.unittest_standard}
Total Time      : ${currentBuild.durationString.minus(' and counting')}```
More info at    : ${env.BUILD_URL}
"""
        message_failure = """*${args.job_status}* CICD Pipeline with details :
```Job             : ${env.JOB_NAME}
Service Name    : ${args.name}
Environment     : ${args.to}
Build Number    : ${env.BUILD_NUMBER}
Service Version : ${args.version}
Unit Test Result: ${args.job_status} with score ${args.unittest_score}/${args.unittest_standard}
Error           : ${args.error}
Stage           : ${args.stage}
Total Time      : ${currentBuild.durationString.minus(' and counting')}```
More info at    : ${env.BUILD_URL}
"""
    } else {
        message_success = """*${args.job_status}* CICD Pipeline with details :
```Job             : ${env.JOB_NAME}
Service Name    : ${args.name}
Environment     : ${args.to}
Build Number    : ${env.BUILD_NUMBER}
Service Version : ${args.version}
Commit Author   : ${args.gitcommit_an}
Commit ID       : ${args.gitcommit_id}
Total Time      : ${currentBuild.durationString.minus(' and counting')}```
More info at    : ${env.BUILD_URL}
"""
        message_failure = """*${args.job_status}* CICD Pipeline with details :
```Job             : ${env.JOB_NAME}
Service Name    : ${args.name}
Environment     : ${args.to}
Build Number    : ${env.BUILD_NUMBER}
Service Version : ${args.version}
Error           : ${args.error}
Stage           : ${args.stage}
Total Time      : ${currentBuild.durationString.minus(' and counting')}```
More info at    : ${env.BUILD_URL}
"""
    }
    if ("${args.job_status}" == 'START') {
        color = 'GREY'
        colorCode = '#D4DADF'
        message = message_start
    } else if ("${args.job_status}" == 'SUCCESS') {
        color = 'GREEN'
        colorCode = '#00FF00'
        message = message_success
    } else if ("${args.job_status}" == 'UNSTABLE') {
        color = 'YELLOW'
        colorCode = '#FFFF00'
        message = message_success
    } else {
        color = 'RED'
        colorCode = '#FF0000'
        message = message_failure
    }
    slackSend color: colorCode, message: message,channel: "${args.channel}"
    googlechatnotification(message: message, url: "${args.gchat}")
}

def sonarScan(Map args) {
    sh "${args.scannerHome}/bin/sonar-scanner -X \
    -Dsonar.projectName=${args.project_name}\
    -Dsonar.projectKey=${args.project_key}\
    -Dsonar.projectVersion=${args.project_version}\
    -Dsonar.sources=. \
    -Dsonar.sources.inclusions=**/**.go \
    -Dsonar.tests=. \
    -Dsonar.test.inclusions=**/**_test.go \
    -Dsonar.test.exclusions=**/vendor/** \
    -Dsonar.go.coverage.reportPaths=coverage.out \
    -Dsonar.go.tests.reportPaths=test-report-${args.name}-${args.project_version}.xml"
    sh "rm -rf test-report-${args.name}-${args.project_version}.xml"
}

def sshSetup() {
    sh "apt-get install sshpass -y > /dev/null 2>&1"
    sh "DEBIAN_FRONTEND=noninteractive apt-get install openssh-server -y > /dev/null 2>&1"
}

def unitTest(Map args) {
    sh "go test ./... -cover -v -covermode=count -coverprofile=coverage.out 2>&1 | go-junit-report > test-report-${args.name}-${args.version}.xml"
    sh "go tool cover -func=coverage.out"
}

def unitTest_publish() {
    junit '*.xml'
}

def unitTest_result() {
    sh(script: 'go tool cover -func=coverage.out | grep total | sed "s/[[:blank:]]*$//;s/.*[[:blank:]]//"', returnStdout: true).trim()
}