import QtQuick 2.4
import QtLocation 5.3

MapQuickItem {
    id: item
    width: 128
    height: 128
    anchorPoint.x: width  * 0.5
    anchorPoint.y: height

    property alias source: marker.source
    property alias logo: logo.source

    sourceItem: Item {
        width: item.width
        height: item.height

        Image {
            id: marker
            anchors.fill: parent
            sourceSize.width: item.width
            sourceSize.height: item.height
        }

        Image {
            id: logo
            anchors.fill: parent
            anchors.leftMargin: item.width * 0.22
            anchors.rightMargin: item.width * 0.22
            anchors.topMargin: item.width * 0.07
            anchors.bottomMargin: item.width * 0.36
            sourceSize.width: item.width * 0.8
            sourceSize.height: item.width * 0.8
        }
    }
}

