### Developer Documentation: Map File Usage

In the `gridmap` package, map files are utilized to define the structure, content, and properties of a game map. Each map is stored in a dedicated directory under a base directory (`mapBaseDir`), and several specific files within this directory are used to populate various elements of the map:

1. **`tileSet.rec`**: This file contains definitions of all possible tiles that can be used in the map. Each tile includes visual information (such as an icon) and properties like walkability and transparency. The tile set is loaded and used to interpret the `tiles.bin` file.

2. **`tiles.bin`**: A binary file that stores the layout of tiles across the grid, mapping each tile position to an index in the `tileSet.rec` file. This effectively lays out the terrain of the map.

3. **`actors.rec`**: This file lists all the actors (e.g., characters, NPCs) present on the map. Each actor's position and attributes are defined, and the file is processed using the provided `actorFactory` to instantiate actors within the grid.

4. **`objects.rec`**: This file specifies the objects placed on the map, including both interactive objects and static features. The `objectFactory` processes each record to generate the appropriate objects and place them on the map.

5. **`items.rec`**: This file contains records of items scattered across the map, detailing their positions and properties. The `itemFactory` is used to add these items to the grid.

6. **`initFlags.rec`** (optional): If present, this file stores initialization flags that can influence map behavior or conditions. These flags are loaded into a map of string keys to integer values and can be used to configure specific map settings.

7. **`iconsForObjects.rec`**: This file maps object categories to their corresponding visual icons. It helps to define the appearance of various objects on the map and is converted into a set of `TextIcon` representations.

These files work together to fully define the visual and interactive elements of a map, allowing the `RecMapLoader` to construct a complete in-game environment.
