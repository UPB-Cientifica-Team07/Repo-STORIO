plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.protobuf)
}

android {
    namespace = "co.edu.upb.cientifica.sync"

    compileSdk = 35

    defaultConfig {
        applicationId = "co.edu.upb.cientifica.sync"

        minSdk = 26
        targetSdk = 35

        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner =
            "androidx.test.runner.AndroidJUnitRunner"
    }

    buildTypes {
        release {
            isMinifyEnabled = false

            proguardFiles(
                getDefaultProguardFile(
                    "proguard-android-optimize.txt"
                ),
                "proguard-rules.pro"
            )
        }
    }

    compileOptions {
        sourceCompatibility =
            JavaVersion.VERSION_11

        targetCompatibility =
            JavaVersion.VERSION_11
    }

    kotlinOptions {
        jvmTarget = "11"
    }
}

protobuf {
    protoc {
        artifact =
            "com.google.protobuf:protoc:3.25.5"
    }

    plugins {
        create("grpc") {
            artifact =
                "io.grpc:protoc-gen-grpc-java:1.68.1"
        }
    }

    generateProtoTasks {
        all().configureEach {
            builtins {
                create("java") {
                    option("lite")
                }
            }

            plugins {
                create("grpc") {
                    option("lite")
                }
            }
        }
    }
}

dependencies {
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.appcompat)
    implementation(libs.material)
    implementation(libs.androidx.activity)
    implementation(libs.androidx.constraintlayout)

    // WorkManager para File Sync en segundo plano.
    implementation("androidx.work:work-runtime-ktx:2.9.1")

    // gRPC Android
    implementation(libs.grpc.okhttp)
    implementation(libs.grpc.protobuf.lite)
    implementation(libs.grpc.stub)

    // Protocol Buffers Lite
    implementation(libs.protobuf.javalite)

    // Generated gRPC sources use javax.annotation.Generated.
    compileOnly(libs.javax.annotation)

    testImplementation(libs.junit)

    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.espresso.core)
}
