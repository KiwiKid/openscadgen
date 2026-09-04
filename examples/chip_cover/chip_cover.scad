include <BOSL2/std.scad>;

$fn = 100;

mode = "obj"; // "obj" "sliced"

chipFloorSize = 0.8;

chipSideSize = 1.5;
 chipWidth = 42; // 
 chipLength = 60; 
 
 chipHeight =  25;

chipSize = [chipLength-3,chipWidth+chipSideSize,chipHeight-1];
cutoutRadius = 6;
chipCutoutSize = [chipLength,chipWidth,chipHeight];

holeCount = 4;
cutoutCirclesUp = chipHeight-cutoutRadius+1;


railHeight = 3;
railUp = 8;
railDepth = 3;
railChamfer = 3;
railInnerChamfer = 1;

module rail(){
up(railUp)
rotate([90,90,0])
    difference(){
        cuboid([railHeight, chipSize[0]-10,chipSize[1]-chipSideSize], chamfer=-railChamfer,);
        cuboid([ railHeight+0.01, chipSize[0]+100,chipSize[1]-railDepth-chipSideSize], chamfer=railInnerChamfer);
        }
        
}

module cordHole(cordHoleHeight=200, cordHoleRadius=cutoutRadius){
ycyl(h=cordHoleHeight, r=cordHoleRadius, $fn=6, anchor=CENTER);
}

module cordHoles(holeOffset=chipSize[0]){

    up(cutoutCirclesUp)
    xcopies(holeOffset/holeCount, holeCount){
    cordHole();

    }
}

module chip_cover(chipFloorSize=chipFloorSize){

chipCutoutMove = [0,0,chipFloorSize];
difference(){
	cuboid(chipSize, anchor=BOTTOM);
    
    move(chipCutoutMove)
    cuboid(chipCutoutSize, anchor=BOTTOM);
    
    cordHoles();
    }
   rail();
}

downShiftOffset = 1.7;
wallHeight = 25;
wallWidth = 100;

wallHoleFloorHeight = 0;
wallHolesAdjustDown = 120 - wallHoleFloorHeight;

holeDiameter = cutoutRadius * 1.1;

module cordHolderWall() {
    difference() {
        cuboid([wallWidth, 1, wallHeight], anchor=BOTTOM);
        
        down(wallHolesAdjustDown) {
            // Vertical copies (Z-axis)
            zcopies(spacing=holeDiameter * 1.5, n=4) {
                // Check row index: shift even rows by 1 diameter along X
                let(rowOffset = ($idx % 2 == 0) ? 0 : holeDiameter)
                right(rowOffset)
                
                // Horizontal copies (X-axis) per row
                xcopies(spacing=holeDiameter * 2.5, n=2)
                    cordHoles(holeOffset=wallWidth);
            }
        }
    }
}


cordHoleOffset = 5;
plugHeight1 = 1;
plugHeight2 = 4;
plugHeight3 = 1;
plugCordHolesCutIn = -2;
module plug(){

difference(){
union(){
    cordHole(cordHoleHeight=plugHeight1);
    
    fwd(plugHeight1/2+plugHeight2/2)
    scale(1.01)
    cordHole(cordHoleHeight=plugHeight2, cordHoleRadius=10);
    
    back(plugHeight1/2+plugHeight3/2)
    scale(1.001)
    #cordHole(cordHoleHeight=plugHeight3, cordHoleRadius=6.3);
    }
    
    #fwd(plugHeight1/2+plugHeight3/2-plugCordHolesCutIn){
    left(cordHoleOffset)
    cyl(r=2, h=100);
    right(cordHoleOffset)
    cyl(r=2, h=100);
    }
    }
}


/*
 chipWidth = 43; // 
 chipLength = 60; 
 
 chipHeight =  30;
 */
wallOnlySectionMove = [10,chipWidth/2,chipHeight-12];
wallOnlySection = [100,1,50];

if(mode == "sliced"){
    intersection(){
    chip_cover();
    cuboid([5,1000,1000]);
    }
}else if(mode == "wallOnly"){
intersection(){
    chip_cover();
    move(wallOnlySectionMove)
    #cuboid(wallOnlySection, anchor=BOT);
    }
} else if(mode == "plug"){
    plug();

} else {
chip_cover();
}



