

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


    jugCutoutDiameter = 37;
    jugCutoutDown = 6;
    bandThickness = 4;
    
    bandWidth = 45;
    bandHeight = 50;
    
    cutoutSize = [jugCutoutDiameter-10,50,700];
    
	module jug_button_cover(){
    
        
        difference(){
        
            cuboid([bandWidth,bandHeight,bandThickness], rounding=2, edges="Z");
            
            
            fwd(jugCutoutDown){
            cyl(d=jugCutoutDiameter, h=1000);
            
            #cuboid([jugCutoutDiameter,30,40], rounding=10);
            
            
            fwd(bandHeight/2)
            cuboid(cutoutSize);
            
            
            up(1)
            back(bandHeight/2-1)
            text3d("Hot", anchor=BOTTOM, height=10, size=6);
            }
            
            }
	}


    sliced(renderType=renderType) {
        jug_button_cover();
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

