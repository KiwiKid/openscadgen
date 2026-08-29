

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

phoneAngle = 80;
phoneHeight = 110;
phoneWidth = 13.5;

baseSize = [65, 65, 20];

phoneMove = [0, 0, phoneHeight/2-1];


phoneAngle2 = 70;
phoneMove2 = [0, 0, phoneHeight/2-4];


phoneAngle3 = 60;
phoneMove3 = [0, -10, phoneHeight/2-7];


phoneAngle4 = 50;
phoneMove4 = [20, 0, phoneHeight/2-13];

phoneRouning =4;

module frame(){
frameInnerWidth = 30;
frameHeight = 12;
    
    up(frameHeight)
    difference(){
        cuboid([100,100,10], anchor=BOTTOM);
        
        cuboid([frameInnerWidth,frameInnerWidth,frameInnerWidth], anchor=BOTTOM);
    }
}
    
	module phone_holder(){
    
    difference(){
		cuboid(baseSize, anchor=BOTTOM, rounding=2);
        
        phoneRatio = 1.08;
        
        move(phoneMove)
        rotate([phoneAngle, 0, 0])
        #cuboid([baseSize[0]*phoneRatio,phoneHeight,phoneWidth], anchor=BOTTOM, rounding=phoneRouning);
        
        move(phoneMove2)
        rotate([phoneAngle2, 0, 90])
        cuboid([baseSize[0]*phoneRatio,phoneHeight,phoneWidth], anchor=BOTTOM, rounding=phoneRouning);
        
         move(phoneMove3)
        rotate([phoneAngle3, 0, 180])
        cuboid([baseSize[0]*phoneRatio,phoneHeight,phoneWidth], anchor=BOTTOM, rounding=phoneRouning);
        
           move(phoneMove4)
        rotate([phoneAngle4, 0, 270])
        cuboid([baseSize[0]*phoneRatio,phoneHeight,phoneWidth], anchor=BOTTOM, rounding=phoneRouning);
        
        frame();
        
      }
	}


    sliced(renderType=renderType) {
        phone_holder();
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
                cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
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

