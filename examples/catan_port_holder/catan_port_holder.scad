

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

    dovetailType = "female";
    dovetailWidth = 13.8;
    dovetailHeight = 5.2;
    
    holderHeight = 3;


portHolderWidth = 30;
    portWidth = 24.5;
    
    holderDepth = 15;
    holderOffset = 0;
    
    dovetailAngle = 34;


	module catan_port_holder(){
        difference(){
        
        cuboid([holderDepth,portHolderWidth,holderHeight], anchor=LEFT);
        
        if(dovetailType == "female"){
            rotate([90,0,270])
            dovetail(dovetailType, slide=100, width=dovetailWidth, height=dovetailHeight, angle=dovetailAngle);
        
        }
        
        move([holderDepth-holderOffset,0,0])
        #cuboid([holderDepth,portWidth,10]);
        }
         if(dovetailType == "male"){
         rotate([90,0,270])
            dovetail(dovetailType, slide=holderHeight, width=dovetailWidth, height=dovetailHeight, angle=dovetailAngle);
            }
	}


    sliced(renderType=renderType) {
        catan_port_holder();
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

