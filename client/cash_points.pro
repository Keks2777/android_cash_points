TEMPLATE = app

QT += qml quick location sql svg

android: QT += androidextras

SOURCES += src/main.cpp \
    src/banklistsqlmodel.cpp \
    src/townlistsqlmodel.cpp \
    src/serverapi.cpp \
    src/listsqlmodel.cpp \
    src/icoimageprovider.cpp \
    src/searchengine.cpp \
    src/locationservice.cpp \
    src/cashpointsqlmodel.cpp \
    src/requests/cashpointrequest.cpp \
    src/requests/cashpointinradius.cpp \
    src/requests/requestfactory.cpp \
    src/requests/cashpointrequestinradiusfactory.cpp

RESOURCES += qml.qrc

# Additional import path used to resolve QML modules in Qt Creator's code model
QML_IMPORT_PATH =

# Default rules for deployment.
include(deployment.pri)

DISTFILES += \
    android/AndroidManifest.xml \
    android/gradle/wrapper/gradle-wrapper.jar \
    android/gradlew \
    android/res/values/libs.xml \
    android/build.gradle \
    android/gradle/wrapper/gradle-wrapper.properties \
    android/gradlew.bat

ANDROID_PACKAGE_SOURCE_DIR = $$PWD/android

HEADERS += \
    src/banklistsqlmodel.h \
    src/townlistsqlmodel.h \
    src/serverapi.h \
    src/rpctype.h \
    src/listsqlmodel.h \
    src/icoimageprovider.h \
    src/searchengine.h \
    src/locationservice.h \
    src/cashpointsqlmodel.h \
    src/requests/cashpointrequest.h \
    src/requests/cashpointinradius.h \
    src/requests/requestfactory.h \
    src/requests/cashpointrequestinradiusfactory.h \
    src/serverapi_fwd.h

QMAKE_CXXFLAGS += -std=c++11

debug {
    DEFINES += CP_DEBUG
}
