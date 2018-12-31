#!/usr/bin/env groovy
def compose_version = 'v0.12.4'
def cli_version = 'v0.5.0'

def namespace = 'decathlon'
def image = 'oauth2_proxy'

node('DOCKER') {
    stage('checkout') {
        checkout scm
    }
    stage(name: 'build Docker') {
      sh "docker build -f docker/Dockerfile -t ${namespace}/${image}:${env.BUILD_NUMBER} ."
    }

    stage('publish docker') {
      if (env.BRANCH_NAME == 'master') {
        docker.withRegistry('https://registry-eu-local.subsidia.org', 'docked_preprod') {
            def dockerImage = docker.image("${namespace}/${image}:${env.BUILD_NUMBER}")
            dockerImage.push "${env.BUILD_NUMBER}"
        }
      }
    }
}
