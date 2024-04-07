package util

import "math"

type CardinalDirection int

const (
	North CardinalDirection = iota
	East
	South
	West
)

type HitInfo2D struct {
	Distance             float64           // Length of the ray until the collision point
	HitSide              CardinalDirection // The side of the obstacle that was hit
	CollisionPosition    [2]float64        // The exact coordinates of the collision point
	ColliderGridPosition [2]int64          // The grid position of the obstacle that was hit
	LastFreeGridPosition [2]int64          // The last free grid position before the collision
	WallPart             float64           // The part of the wall that was hit (0.0 (left-most) to 1.0 (right-most))
	TraversedGridCells   [][2]int64
	Origin               [2]int64
}

// Raycast2D casts a ray from the given start position into the given direction.
// Direction is a NORMALIZED vector.
func Raycast2D(startX, startY, directionX, directionY float64, shouldStopRay func(x, y int64) bool) HitInfo2D {
	// 2. Die sich wiederholenden Strahlenteile für Schritte in CollisionX und CollisionY Richtung berechnen
	//    - Steigung des Richtungsvektors und ein Schritt in CollisionX bzw. CollisionY Richtung
	//    - Satz des Pythagoras mit a = 1 und b = ray.y / ray.x für einen Schritt in die CollisionX-Richtung
	//    - Satz des Pythagoras mit a = 1 und b = ray.x / ray.y für einen Schritt in die CollisionY-Richtung
	deltaDistX := math.Sqrt(1 + (directionY*directionY)/(directionX*directionX))
	deltaDistY := math.Sqrt(1 + (directionX*directionX)/(directionY*directionY))

	startGridX := int64(math.Floor(startX))
	startGridY := int64(math.Floor(startY))

	mapX := startGridX
	mapY := startGridY

	var intraCellPositionX float64
	var intraCellPositionY float64

	var sideDistX float64
	var sideDistY float64

	var mapStepX int64
	var mapStepY int64

	var traversedGridCells [][2]int64
	var colliderGridPosition [2]int64

	traversedGridCells = append(traversedGridCells, [2]int64{mapX, mapY})
	// 3. Anhand der Richtung des Strahls folgende Werte berechnen
	//     - Die Schrittrichtung auf der Karte (je Quadrant)
	//     - Die translation innerhalb der aktuellen Zelle
	//     - Die anteiligen Startteile der beiden Strahlen für CollisionX und CollisionY Schritte
	if directionX < 0 {
		mapStepX = -1
		intraCellPositionX = startX - float64(mapX)
		sideDistX = intraCellPositionX * deltaDistX
	} else {
		mapStepX = 1
		intraCellPositionX = float64(mapX) + 1.0 - startX
		sideDistX = intraCellPositionX * deltaDistX
	}

	if directionY < 0 {
		mapStepY = -1
		intraCellPositionY = startY - float64(mapY)
		sideDistY = intraCellPositionY * deltaDistY
	} else {
		mapStepY = 1
		intraCellPositionY = float64(mapY) + 1.0 - startY
		sideDistY = intraCellPositionY * deltaDistY
	}

	eastWestSide := false
	stopRaycasting := false
	nextCollisionX := 0.0
	nextCollisionY := 0.0

	// 4. Schrittweise die Strahlen verlängern, immer der kürzeste zuerst
	//    - Die Kartenposition wird aktualisiert
	//    - Die Wandrichtung (Nord/Süd vs. Ost/west) wird gesetzt
	//    - Der Strahl wird verlängert
	for !stopRaycasting {
		if sideDistX < sideDistY {
			// move one unit in CollisionX Direction
			nextCollisionX = startX + (directionX * sideDistX)
			nextCollisionY = startY + (directionY * sideDistX)

			mapX += mapStepX
			eastWestSide = true
			sideDistX += deltaDistX
		} else {
			// move one unit in CollisionY Direction
			nextCollisionX = startX + (directionX * sideDistY)
			nextCollisionY = startY + (directionY * sideDistY)

			mapY += mapStepY
			eastWestSide = false
			sideDistY += deltaDistY
		}

		// 5. Prüfen wir ob eine Wand getroffen wurde
		stopRaycasting = shouldStopRay(mapX, mapY)
		if stopRaycasting {
			colliderGridPosition = [2]int64{mapX, mapY}
		} else {
			traversedGridCells = append(traversedGridCells, [2]int64{mapX, mapY})
		}
	}

	// 6. Den Abstand zur Wand berechnen (Im Prinzip den letzten Schritt rückgängig machen)
	// Vektoren subtrahieren
	var perpWallDist float64
	if eastWestSide {
		perpWallDist = sideDistX - deltaDistX
	} else {
		perpWallDist = sideDistY - deltaDistY
	}

	// 6.1 Die original Distanz zur Wand für das Texture Mapping verwenden
	var wallX float64 // CollisionX translation an der wir die Wand getroffen haben
	// Komponentenweise Vektoraddition und Skalierung
	if eastWestSide {
		wallX = startY + (perpWallDist * directionY)
	} else {
		wallX = startX + (perpWallDist * directionX)
	}

	wallX -= math.Floor(wallX) // Uns interessieren nur die Nachkommastellen
	//textureX := (int)(wallX * mWallTextures[textureIndex].Width);

	var hitSide CardinalDirection
	if eastWestSide {
		if directionX < 0 {
			hitSide = West
		} else {
			hitSide = East
		}
	} else {
		if directionY < 0 {
			hitSide = North
		} else {
			hitSide = South
		}
	}
	return HitInfo2D{
		Distance:             perpWallDist,
		HitSide:              hitSide,
		CollisionPosition:    [2]float64{nextCollisionX, nextCollisionY},
		LastFreeGridPosition: traversedGridCells[len(traversedGridCells)-1],
		WallPart:             wallX,
		TraversedGridCells:   traversedGridCells,
		ColliderGridPosition: colliderGridPosition,
		Origin:               [2]int64{startGridX, startGridY},
	}
}

func ReflectingRaycast2D(startX, startY, startDirectionX, startDirectionY float64, allowedReflectionCount int, shouldReflectRay func(x, y int64) bool) []HitInfo2D {
	var result []HitInfo2D

	originX := startX
	originY := startY
	directionX := startDirectionX
	directionY := startDirectionY

	for i := 0; i < allowedReflectionCount; i++ {
		rayCast := Raycast2D(originX, originY, directionX, directionY, shouldReflectRay)
		result = append(result, rayCast)
		originX = rayCast.CollisionPosition[0]
		originY = rayCast.CollisionPosition[1]

		if rayCast.HitSide == North || rayCast.HitSide == South {
			directionY = -directionY
		} else {
			directionX = -directionX
		}
	}

	return result
}

func ChainedRaycast2D(startX, startY, startDirectionX, startDirectionY float64, isEndOfOneRay func(x, y int64) bool, nextTarget func(x, y int64) (bool, int, int)) []HitInfo2D {
	var result []HitInfo2D

	originX := startX
	originY := startY
	directionX := startDirectionX
	directionY := startDirectionY
	maxChain := 100
	for i := 0; i < maxChain; i++ {
		rayCast := Raycast2D(originX, originY, directionX, directionY, isEndOfOneRay)
		result = append(result, rayCast)
		originX = float64(rayCast.ColliderGridPosition[0]) + 0.5
		originY = float64(rayCast.ColliderGridPosition[1]) + 0.5

		if shouldContinue, nextX, nextY := nextTarget(rayCast.ColliderGridPosition[0], rayCast.ColliderGridPosition[1]); shouldContinue {
			directionX = float64(nextX) - originX
			directionY = float64(nextY) - originY
			length := math.Sqrt(directionX*directionX + directionY*directionY)
			directionX /= length
			directionY /= length
		} else {
			break
		}
	}

	return result
}
