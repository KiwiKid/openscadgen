include <BOSL2/std.scad>;

$fn = 100;

mode = "hexTable";// "pipeSize" "withSimulatedPipeLegs"

tableSize = [20,20,3];

pvcPipeRadius = 2.6;
pvcPipeWallSize = 0.4;
pvcPipeHoleHeight = 12.8;
holeOutAdjust = 7;

module legs(mode="holes"){
    zrot(8)
       rot_copies(n=3)
       xrot(-15)
       back(holeOutAdjust)
       pvcCutoutHole();

}

module pvcCutoutHole(tubeHeight =pvcPipeHoleHeight){
	#tube(or=pvcPipeRadius,
        wall=pvcPipeWallSize,
        h=tubeHeight,
        anchor=BOT,
     //   chamfer2=-0.1,
    //    ichamfer2=-.3
        );

}

module small_pvc_stand(){
intersection(){
    cuboid(tableSize, rounding=2, edges="Z", anchor=TOP);
    cuboid(tableSize, rounding=1, anchor=TOP);
    }
}

if(mode == "pipeSize"){

    difference(){
    cuboid([20,20,3]);
pvcCutoutHole();
}

}else if(mode == "hexTable") {
    difference(){
    small_pvc_stand();
    
legs();
        
    }
} else if(mode == "withSimulatedPipeLegs"){
small_pvc_stand();

} else {

}
