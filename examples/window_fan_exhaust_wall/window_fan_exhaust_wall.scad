

	include <BOSL2/std.scad>;
include <BOSL2/joiners.scad>;
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
    
    
        fanDiameter = 40;
        fanOffset = 10;

        cubeWidth = 100;
        cubeHeight = 20;
        cubeDepth = 110;
        
        hasFanHole = false; // false
        hasStartDovetail = true;
        hasEndDovetail = true;
        dovetailSlideAllowance = 0.2;
        
        
        endType = "sideHolder";
         isOnSide= true;
        sideHolderOffset = 10;
        


    module fan_hole(fanDiameter=fanDiameter){
        cyl(50,fanDiameter, center=true);
   }

module end_dovetail(isFemale = true){
sideWidth = 15;
    if (isFemale) {
    
        if(isOnSide == true){
            rotate([270,0,0])
            move([cubeWidth/2-sideHolderOffset,0,-cubeHeight/2])
            dovetail("female", slide=cubeDepth, width=sideWidth, height=8);
        } else {
            rotate([90,0,90])
        move([0,0,cubeWidth/2])
        dovetail("female", slide=cubeDepth, width=sideWidth, height=8);
        }
    } else {
        rotate([90,0,270])
        move([0,0,cubeWidth/2])
        dovetail("male", slide=cubeDepth, width=sideWidth-dovetailSlideAllowance, height=8);
    }
}

	module window_fan_exhaust_wall(){


        difference() {
        
            		cuboid([cubeWidth,cubeHeight,cubeDepth], rounding=1, edges=[TOP,BOTTOM]);

        if(hasStartDovetail == true && isOnSide == false){
            end_dovetail(isFemale=true);
        }

        

        if(hasFanHole == true){
            rotate([90,0,0])
            fan_hole(fanDiameter=fanDiameter);
        }
}


        if(hasStartDovetail == true && isOnSide == true){
            end_dovetail(isFemale=true);
        }
          
                if(hasEndDovetail == true){
                    end_dovetail(isFemale=false);
                }

        
	}


    sliced(renderType=renderType) {
        window_fan_exhaust_wall();
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

