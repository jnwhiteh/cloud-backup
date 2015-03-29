﻿# Location facet

The **Location** facet groups geographic location-related data on OneDrive into a single structure.

It is available on the location property of [Item][item-resource] resources that have
an associated geographic location.

## JSON representation

<!-- { "blockType": "resource", "@odata.type": "oneDrive.location" } -->
```json
{
  "altitude": 760.0,
  "latitude": 122.1232,
  "longitude": 34.0012
}
```

## Properties
| Property name | Type   | Description                                                    |
|:--------------|:-------|:---------------------------------------------------------------|
| **altitude**  | number | The altitude (height), in feet,  above sea level for the item. |
| **latitude**  | number | The latitude, in decimal, for the item.                       |
| **longitude** | number | The longitude, in decimal, for the item.                      |


[item-resource]: ../resources/item.md
