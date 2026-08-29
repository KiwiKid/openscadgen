include <BOSL2/std.scad>;
include <BOSL2/screws.scad>;

$fn = 200;



holderHeight = 40;
baseRadius   = 70;    // Sturdy base to prevent tipping
topBaseRadus     = 38; 

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

module can_tip_preventer(){
    difference(){
        cyl(r1=baseRadius, r2=topBaseRadus, h=holderHeight, rounding=2, anchor=BOT);
        
       // #down(holderHeight/2+rimHeight/2+20)
        up(floorHeight)
 //       down(holderHeight)
    union(){
        up(bigCanFloorOffset)
        cyl(r=canRadius, h=holderHeight+rimHeight, anchor=BOT);
       cyl(r=canRadius, h=holderHeight+rimHeight, anchor=BOT,chamfer1=10);
       
       cyl(r=slimCanRadius, h=holderHeight+rimHeight, anchor=BOT);
       }
       
       

    }
    
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