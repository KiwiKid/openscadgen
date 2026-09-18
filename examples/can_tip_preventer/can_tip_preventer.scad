include <BOSL2/std.scad>;
include <BOSL2/screws.scad>;

$fn = 200;



holderHeight = 40;
baseRadius   = 70;    // Sturdy base to prevent tipping
topBaseRadus     = 38; 

duplicateCountX = 2;
duplicateRotateSkew = 10;

duplicateCountDistance = 100;
duplicateCountBodyDistance = 60;

duplicateCountY = 2;


canRadiusWiggle = 2;
standardCanRadius = 33;
canRadius = standardCanRadius+canRadiusWiggle;  // 33mm can radius
canHeight    = 156;   // Standard 12oz can height

slimCanRadius = 28.5; 
slimCanHeight = 156;

floorHeight  = 2;
bigCanFloorOffset = 10;
rimHeight    = canHeight*0.4;
rimWallSize  = 2;

sidePocketDepth = 3;
sidePocketOut = canRadius+rimWallSize+10;
sidePocketSize = [20,35,200];
sidePocketSizePhone = [20,100,200];

ridgeDepth = 5;

grapGapUp = 65;
grapGapWidth = 25;

holderMode = "bottle";

bottleBottomRadius = 34;
bottleTopRadius = 40;

module can_tip_preventer_body(holderMode="can"){
    if(holderMode == "can"){
        cyl(r1=baseRadius, r2=topBaseRadus, h=holderHeight, rounding=2, anchor=BOT);
    } else if(holderMode == "bottle"){
                cyl(r1=baseRadius, r2=topBaseRadus, h=holderHeight, rounding=2, anchor=BOT);

    
   } else {
   
    assert(false, "invalid holderMode can_tip_preventer_body");
    }
}

module can_tip_preventer_cutout(holderMode=holderMode){
    if(holderMode == "can"){
        up(floorHeight)
         //       down(holderHeight)
            union(){
                up(bigCanFloorOffset)
                cyl(r=canRadius, h=holderHeight+rimHeight, anchor=BOT);
               cyl(r=canRadius, h=holderHeight+rimHeight, anchor=BOT,chamfer1=10);
               
               cyl(r=slimCanRadius, h=holderHeight+rimHeight, anchor=BOT);
               }
    } else if(holderMode == "bottle"){
    up(floorHeight)
 //       down(holderHeight)
    union(){
        up(bigCanFloorOffset)
        cyl(r=bottleBottomRadius, h=holderHeight+rimHeight, anchor=BOT);
       #cyl(r1=bottleBottomRadius, r2=bottleTopRadius+2, h=holderHeight+rimHeight, anchor=BOT,chamfer1=10);
       
       cyl(r=slimCanRadius, h=holderHeight+rimHeight, anchor=BOT);
       }

    
   } else {
   
    assert(false, "invalid holderMode in can_tip_preventer_cutout");
    }
}


module can_tip_preventer(holderMode=holderMode){

    difference(){
    
        xcopies(duplicateCountBodyDistance, duplicateCountX)
        ycopies(duplicateCountBodyDistance, duplicateCountY)
        can_tip_preventer_body(holderMode=holderMode);
        
        
        
        
       // #down(holderHeight/2+rimHeight/2+20)
       grid_copies(spacing=duplicateCountDistance, n=[duplicateCountX, duplicateCountY]) {
    
    // 2. Point the object away from the center (0,0) using its grid position
    rot(from=Z, to=Z + normalize([$pos.x, $pos.y, 0]) * 0.2) {
       can_tip_preventer_cutout(holderMode=holderMode);
       }
       }
       

    }
    
    
xcopies(duplicateCountDistance, duplicateCountX)
ycopies(duplicateCountDistance, duplicateCountY)
    up(holderHeight-1)
    difference(){
        cyl(r=canRadius+rimWallSize, h=rimHeight+rimWallSize, anchor=BOT);
        down(0.02)
        cyl(r=canRadius, h=rimHeight+floorHeight+0.03, anchor=BOT, chamfer=-2);
        
        
        up(grapGapUp)
        cuboid([grapGapWidth,300,100], rounding=10);

    }
    
    
    
}
can_tip_preventer();