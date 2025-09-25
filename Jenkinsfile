pipeline {
    agent any

    environment {
        DOCKER_REGISTRY = "your-docker-registry.com" // Replace with your Docker registry
        DOCKER_IMAGE_NAME = "ecommerce-{{ .Values.service.name }}" // Placeholder for service name
        HELM_CHART_PATH = "deployments/kubernetes/charts/{{ .Values.service.name }}" // Placeholder for service chart path
        KUBECONFIG_CREDENTIAL_ID = "your-kubeconfig-credential-id" // Jenkins credential ID for kubeconfig
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Build Go Application') {
            steps {
                script {
                    // Assuming go.mod is in the root, and service main is in cmd/<service-name>/main.go
                    sh "go mod tidy"
                    sh "CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build/{{ .Values.service.name }} ./cmd/{{ .Values.service.name }}"
                }
            }
        }

        stage('Run Tests') {
            steps {
                sh "go test ./..."
            }
        }

        stage('Build Docker Image') {
            steps {
                script {
                    // Build Docker image for the service
                    sh "docker build -t ${DOCKER_REGISTRY}/${DOCKER_IMAGE_NAME}:${env.BUILD_NUMBER} -f Dockerfile ."
                }
            }
        }

        stage('Push Docker Image') {
            steps {
                script {
                    // Push Docker image to registry
                    withCredentials([usernamePassword(credentialsId: 'docker-hub-credentials', passwordVariable: 'DOCKER_PASSWORD', usernameVariable: 'DOCKER_USERNAME')]) {
                        sh "echo \"$DOCKER_PASSWORD\" | docker login ${DOCKER_REGISTRY} -u \"$DOCKER_USERNAME\" --password-stdin"
                        sh "docker push ${DOCKER_REGISTRY}/${DOCKER_IMAGE_NAME}:${env.BUILD_NUMBER}"
                    }
                }
            }
        }

        stage('Deploy to Kubernetes') {
            steps {
                script {
                    // Deploy using Helm
                    withCredentials([file(credentialsId: KUBECONFIG_CREDENTIAL_ID, variable: 'KUBECONFIG_FILE')]) {
                        sh "helm upgrade --install {{ .Values.service.name }} ${HELM_CHART_PATH} --namespace default --set image.tag=${env.BUILD_NUMBER} --kubeconfig ${KUBECONFIG_FILE}"
                    }
                }
            }
        }
    }
}