

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
	renderType = "horzSlice";

ringHeight = 50;

ringWidth = 32;
ringThickness = 0.5;
	module mic_cover(){
    
    module column(radius=12){
    
        fwd(15)
        up(5)
        rotate([90,0,0])
        cyl(r=radius, h=50);
        
        fwd(15)
        rotate([90,0,0])
        cyl(r=radius, h=50);
        
        fwd(15)
        down(5)
        rotate([90,0,0])
        cyl(r=radius, h=50);
        
        
        fwd(15)
        down(10)
        rotate([90,0,0])
        cyl(r=radius, h=50);
        
        
        fwd(15)
        down(15)
        rotate([90,0,0])
        cyl(r=radius, h=50);
        
        fwd(15)
        down(20)
        rotate([90,0,0])
        cyl(r=radius, h=50);
        
        fwd(15)
        down(25)
        rotate([90,0,0])
        cyl(r=radius, h=50);
    }
    difference(){
		cyl(r1=ringWidth, r2=ringWidth, h=ringHeight);
        
        cyl(r1=ringWidth-ringThickness, r2=ringWidth-ringThickness, h =ringHeight+1);
       
       down(5)
       column(radius=16);
       
       up(5)
       rotate([0,0,90])
       back(ringWidth)
       column(radius=10);
        
        
        rotate([90,0,90])
        up(ringWidth-8)
        linear_extrude(height = 10) {
            text("GC", size = 15, font = "Arial", halign = "center", valign = "center");
        }
        
        }
	}


    sliced(renderType=renderType) {
        #mic_cover();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.5,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 20],
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

