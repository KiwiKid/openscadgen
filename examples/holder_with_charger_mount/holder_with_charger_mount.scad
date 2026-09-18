include <BOSL2/std.scad>;
include <BOSL2/isosurface.scad>;
$fn = 100;

mode = "phone";

phoneLeanRotate = 15;

phoneWedgeSize = 70;

phoneCutoutBaseWidth = 87;
phoneCutoutBaseHeight = 90;
phoneCutoutBaseDepth = 14;

phoneCutoutMiddleWidth = 60;
phoneCutoutMiddleHeight = 90;
phoneCutoutMiddleDepth = 4;

phoneCutoutTopWidth = phoneWedgeSize+1;
phoneCutoutTopHeight = 90;
phoneCutoutTopDepth = 50;

phoneMoveDown = 20;

phoneMove = [phoneWedgeSize/2,phoneWedgeSize-phoneMoveDown,-20];


module phoneLander(){


}

module phoneLeanRotate(heightIncrease=0){

   cuboid([phoneCutoutBaseWidth, phoneCutoutBaseHeight+heightIncrease, phoneCutoutBaseDepth], anchor=TOP, rounding=1, edges="Z");
    /*up(phoneCutoutBaseDepth/2-0.001)
         cuboid([phoneCutoutMiddleWidth, phoneCutoutMiddleHeight, phoneCutoutMiddleDepth], anchor=BOT)
         up(phoneCutoutMiddleDepth/2-0.002)
         
         cuboid([phoneCutoutTopWidth, phoneCutoutTopHeight, phoneCutoutTopDepth], anchor=BOT);*/

}

module connector(topDiskPosition=[0,0,50], topDiskRotate=[1,15,0]){
    // 1. Top Disk (Rotated X amount, small/medium radius, high up)
top_r = 15;
top_h = 5;
top_transform = up(60) * rot(topDiskRotate); 

// 2. Middle Disk (Small radius, halfway up)
mid_r = 8;
mid_h = 4;
mid_transform = move(topDiskPosition);

// 3. Ground Disk (Large radius, at the base)
ground_r = 30;
ground_h = 6;
ground_transform = IDENT; // No transformation needed (Identity matrix)

// --- Metaball Generation ---
// The bounding box must safely envelope all shapes + their interaction range
bbox = [
    [-100, -145, -15], 
    [100, 145, 155]
];
voxel_size = 2;
metaballs(
    spec = [
        ground_transform, mb_disk(h=ground_h, r=ground_r),
        mid_transform,    mb_disk(h=mid_h, r=mid_r),
        top_transform,    mb_disk(h=top_h, r=top_r)
    ],
    bounding_box = bbox,
    voxel_size = voxel_size,
    isovalue = .33 // Lower values (e.g. 0.8) increase blending/melting strength
);


}


phoneCutoutMove = [0, 4.5, 1];

module holder_with_charger_mount(mode=mode){
if(mode == "phone"){

connector();


	//wedge(phoneWedgeSize);
    
    rotate([90-phoneLeanRotate,0, 0])
    move(phoneMove)
    difference(){
    scale(1.1)
        phoneLeanRotate();
        move(phoneCutoutMove)
        phoneLeanRotate(heightIncrease=10);
    
        
}

} else if(mode == "ipad"){
    difference(){

        wedge(20)
        attach(TOP){
            cuboid([10,10,10]);
        }
            
    }

    } else {
 
    assert(false, "invalid mode");
    }
}

holder_with_charger_mount();
