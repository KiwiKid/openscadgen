

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";
    partType = "right";

holderSize = [20,35,1];

holderRotate = 70;
legOut =7;

mountObjMove = [0,7,5];
mountObjSize = [60,8,20];

// mountObjRotate = [5,0,0];
mountObjRotate = [30,0,0];


// large-square | small-square
mount = "very-large-square";

smallMountObjSize = [60,4,20];
smallMountObjMove =  [0,6,5];


mountScale = 4;

    module holder_arm(partType=partType){
        if(partType == "all" || partType == "left"){

            cuboid(holderSize, rounding=10, edges="Z", except=[FWD-LEFT]);
        }
        
        if(partType == "all" || partType == "right"){
           
            cuboid(holderSize, rounding=10, edges="Z", except=[FWD-LEFT]);
        
       }
    }


	module simple_holder(){
    
    difference(){
    
    union(){
    
    /* The Holder cutouts */
       if(partType == "all" || partType == "left"){
       
                   left(legOut)
            rotate([0,-50,30])
            holder_arm(partType=partType);
       }
      
      if(partType == "all" || partType == "right"){

          right(legOut)
             rotate([0,50,-30])
            holder_arm(partType=partType);
       }
       };
       
       
       /* The Cutthroughts */
       if(partType == "all" || partType == "right"){
           left(legOut)
           rotate([0,-50,30])
           move([-5,-10,0])
            holder_arm();
       }
       
       
      if(partType == "all" || partType == "left"){
      
          right(legOut)
         rotate([0,50,-30])
           move([-7,5,00])
            holder_arm();
       
       }
       

       
       
       // MOUNT OBJECT TYPE
       if(mount == "large-square"){
       
              rotate(mountObjRotate)
       move(mountObjMove+[0,10,5])
       cuboid(mountObjSize+[0,5,0], rounding=1 );
       
       
           rotate(mountObjRotate)
           move(mountObjMove)
           cuboid(mountObjSize, rounding=1);
       }
       
       if(mount == "very-large-square"){
       
              rotate(mountObjRotate)
       move(mountObjMove+[0,10,5])
       cuboid(mountObjSize+[0,10,0], rounding=1 );
       
       
           rotate(mountObjRotate)
           move(mountObjMove+[0,-2,0])
           cuboid(mountObjSize+[0,3,0], rounding=1);
       }
       
        if(mount == "small-square"){
        
               rotate(mountObjRotate)
           move(mountObjMove+[0,5,3])
           cuboid(mountObjSize+[0,5,0], rounding=1 );
        
           rotate(mountObjRotate)
           move(smallMountObjMove)
           cuboid(smallMountObjSize, rounding=1);
           
       }
       
       
       
       }
	}


    sliced(renderType=renderType) {
    
    scale(mountScale)
        simple_holder();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.2,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cube([sliceThickness, sliceSize, sliceSize], center=false);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

