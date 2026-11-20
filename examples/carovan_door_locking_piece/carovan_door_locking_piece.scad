include <BOSL2/std.scad>;

$fn = 100;

lockingBoxSize = [60,35, 4];
lockingBoxRounding = 2;

boltFlatWidth = 6;
boltThreadRadius = 3.5;
boltHolderHexRadius = 6;

boltOuterRadius = 8;
boltOuterRadiusChamfer1 = -1.5;
boltOuterTubeHeight = 20;

boltHoleHeight = 8;

holeSide = 1; // -1 for left
holeFwd = (-lockingBoxSize[0]/2+boltOuterRadius+-boltOuterRadiusChamfer1)*holeSide;
holeCenterOffset = 3;


wedgeHeight = lockingBoxSize[2]*0.7;
wedgeSize = [lockingBoxSize[0]+1, wedgeHeight, 11];

endStopperWidth = 3;
endStopperHeight = 5;

module thru_hole(mode="flatSideCircle"){
    if(mode=="flatSideCircle"){
        intersection(){
            cyl(r=boltThreadRadius, h=boltOuterTubeHeight*2);
            
            cuboid([boltFlatWidth,100,100]);
        }
    }
}

module carovan_door_locking_piece(){
    difference(){
        union(){
            cuboid(lockingBoxSize, chamfer=lockingBoxRounding)
            attach(BOT)
            fwd(lockingBoxSize[1]/2-endStopperWidth/2)
            cuboid([lockingBoxSize[0],endStopperWidth,endStopperHeight], chamfer=1.3, except_edges=[BOT+LEFT]);
            
            
            up(-lockingBoxSize[2]/2-boltHoleHeight*2)
            right(holeFwd)
            fwd(holeCenterOffset)
            //cyl(r=boltOuterRadius, h=boltOuterTubeHeight, anchor=BOT, chamfer1=boltOuterRadiusChamfer1, chamfer2=1);
            
            
                 linear_extrude(boltOuterTubeHeight)
        hexagon(r=boltOuterRadius, realign=true);
            
        }
        
        fwd(holeCenterOffset)
        right(holeFwd){
            
         //   #up(-boltOuterTubeHeight-boltHoleHeight+0.001)
    //        down(boltOuterTubeHeight)
   //    linear_extrude(boltHoleHeight)
    //   hexagon(ir=boltHolderHexRadius, realign=true);
        
       
           // down(boltHoleHeight)
          down(lockingBoxSize[2]/2+0.001)
    linear_extrude(boltHoleHeight)
        hexagon(ir=boltHolderHexRadius, realign=true);
        
            
        thru_hole(mode="flatSideCircle");
        }
        
       down(lockingBoxSize[2]/2-wedgeHeight/2+0.01)
       // fwd(lockingBoxSize[1]/2-wedgeSize[1]/2)
        back(lockingBoxSize[1]/2-wedgeSize[2]/2+0.01)
        xrot(90)
        wedge(wedgeSize, anchor=CENTER);
        
        
    }
}

carovan_door_locking_piece();
