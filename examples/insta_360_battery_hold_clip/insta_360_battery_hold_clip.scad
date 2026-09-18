include <BOSL2/std.scad>;
include <BOSL2/isosurface.scad>;
$fn = 100;

mode = "case"; // stand
cutMode ="off"; //back

clipThickness = 2;
clipBandHeight= 4; // 10 12
clipRounding = 8; 

wiggleRoom = 0.3;
instaSize = [46.2, 20.8, 113.0]+[wiggleRoom,wiggleRoom,0];
clipSize = [46.2+clipThickness, 20.8+clipThickness, clipBandHeight];

instaLenHeight = 90;
instaLenRadis = 20;

module instaLensShape(r=instaLenRadis){
    sphere(r=instaLenRadis);
}

module insta_shape(){

    cuboid(instaSize, anchor=BOT, rounding=clipRounding, edges="Z");
    up(instaLenHeight)
    instaLensShape(r=instaLenRadis);
    
}

module instaCaseTopTransition(){
spec = [
    move([0,0,20]), mb_sphere(instaLenRadis/2, hide_debug=true),
    move([0,0,20]), mb_cuboid(clipSize)
];
metaballs(spec, voxel_size=0.5
,   bounding_box=[[-30,-20,-12], [15,12,12]]
, debug=true
, $fn=2
);

}

caseMiddleHeight = 80;

module insta_case(){
difference(){
    union(){
	cuboid([clipSize[0],clipSize[1],caseMiddleHeight], anchor=BOT, rounding=clipRounding, edges="Z");
    cuboid([clipSize[0],clipSize[1],10], anchor=TOP, rounding=clipRounding, except_edges=TOP);
    up(instaLenHeight){
   instaLensShape(r=instaLenRadis*1.1);
   
    
    #instaCaseTopTransition();
   }
   
    }
    
    //up(clipThickness)
    down(1)
        insta_shape();
    }


}

module insta_tube(){
difference(){
	cuboid(clipSize, anchor=BOT, rounding=clipRounding, edges="Z");
    
    //up(clipThickness)
    down(1)
        insta_shape();
    }

}

module insta_360_battery_hold_clip(mode="ring"){

    if(mode == "ring"){
        intersection(){
        insta_tube();
        up(0)
            #cuboid([1000, 1000, clipBandHeight]);
            
        }
    } else if(mode == "stand"){
      //  intersection(){
      insta_tube();
     //   up(0)
     //       cuboid([1000, 1000, 1000]);
            
     //   }

    } else if(mode == "case"){
      //  intersection(){
     insta_case();
     //   up(0)
     //       cuboid([1000, 1000, 1000]);
            
     //   }

    }else if(mode == "case-dev-top") {
    instaCaseTopTransition();
//        assert(false, "invalid mode");
    }
}


if(cutMode =="back"){
    back_half()
    insta_360_battery_hold_clip(mode=mode);
 }else {
  insta_360_battery_hold_clip(mode=mode);
 }
