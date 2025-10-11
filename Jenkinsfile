// Jenkinsfile (Declarative Pipeline)

pipeline {
    // agent any 表示流水线可以在任何可用的 agent 上执行
    agent any

    environment {
        // Docker 仓库地址
        DOCKER_REGISTRY = "your-docker-registry"
        // Docker 凭证 ID (在 Jenkins 中配置)
        DOCKER_CREDENTIALS_ID = "your-docker-credentials"
        // Kubernetes 配置文件凭证 ID (在 Jenkins 中配置)
        KUBECONFIG_CREDENTIALS_ID = "your-kubeconfig-credentials"
        // 部署的目标命名空间
        NAMESPACE = "default"
    }

    parameters {
        // 定义一个参数，用于在运行时指定要构建的服务名称
        string(name: 'SERVICE_NAME', defaultValue: 'user', description: 'The name of the microservice to build and deploy (e.g., user, product, order).')
    }

    stages {
        // 阶段 1: 代码检出
        stage('Checkout') {
            steps {
                // 从版本控制系统拉取代码
                git branch: 'main', url: 'https://github.com/your-repo/ecommerce.git'
            }
        }

        // 阶段 2: 生成 Protobuf 代码
        stage('Generate Proto') {
            steps {
                // 确保脚本有执行权限
                sh 'chmod +x ./scripts/gen_proto.sh'
                // 执行脚本生成代码
                sh './scripts/gen_proto.sh'
            }
        }

        // 阶段 3: 单元测试
        stage('Unit Test') {
            steps {
                // 在 Go 容器中执行测试
                // 注意: 确保你的测试代码和依赖是完整的
                sh 'docker run --rm -v $(pwd):/app -w /app golang:1.22 go test -v ./...'
            }
        }

        // 阶段 4: 构建并推送 Docker 镜像
        stage('Build & Push Docker Image') {
            steps {
                script {
                    // 定义镜像标签，使用 Jenkins 的 BUILD_ID 保证唯一性
                    def imageTag = "${env.BUILD_ID}"
                    def serviceName = params.SERVICE_NAME
                    def imageName = "${DOCKER_REGISTRY}/${serviceName}:${imageTag}"

                    // 构建镜像，通过 build-arg 传递服务名称
                    def dockerImage = docker.build(imageName, "--build-arg SERVICE_NAME=${serviceName} -f Dockerfile .")

                    // 使用 Jenkins 的 Docker 凭证插件进行推送
                    docker.withRegistry("https://${DOCKER_REGISTRY}", DOCKER_CREDENTIALS_ID) {
                        dockerImage.push()
                    }
                }
            }
        }

        // 阶段 5: 部署到 Kubernetes
        stage('Deploy to Kubernetes') {
            steps {
                script {
                    def imageTag = "${env.BUILD_ID}"
                    def serviceName = params.SERVICE_NAME

                    // 使用 withKubeConfig 插件来安全地使用 kubeconfig 文件
                    withKubeConfig([credentialsId: KUBECONFIG_CREDENTIALS_ID]) {
                        // 使用 Helm 来部署或升级应用
                        // 通过 --set 参数覆盖 values.yaml 中的镜像版本
                        sh """helm upgrade --install ${serviceName} \
                           ./deployments/kubernetes/charts/${serviceName} \
                           --namespace ${NAMESPACE} \
                           --set image.repository=${DOCKER_REGISTRY}/${serviceName} \
                           --set image.tag=${imageTag} \
                           --create-namespace"""
                    }
                }
            }
        }
    }

    // post 块定义了流水线完成后执行的操作
    post {
        always {
            // 清理工作区
            cleanWs()
        }
    }
}