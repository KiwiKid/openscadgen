

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

    module cylSegment(){
        difference(){
        cyl(r=4);
        cyl(r=3, 2);
        }
    }
    
    wallDepth = 30;
    wallSize = 2;
    topBoxWidth = 35;
    topBoxHeight = 32;
    bottomBoxWidth = 44; 
    bottomBoxHeight = 36; 
    module squareWithGap(height=10, width=10){
    difference(){
        cuboid([wallDepth,width,height]);
        cuboid([wallDepth+1,width-wallSize,height-wallSize]);
       }
    }
    
	module sink_hose_extender(){

        
        up(bottomBoxHeight/2)
        squareWithGap(height=topBoxHeight, width=topBoxWidth);
        down(topBoxHeight/2)
        union(){
        squareWithGap(height=bottomBoxHeight, width=bottomBoxWidth);
        up(bottomBoxHeight/4)
        right(wallDepth/2-3)
         cuboid([5, bottomBoxWidth*0.7, bottomBoxHeight/2 ], rounding=2, edges=[BOT]);
        }
        
        
	}


    sliced(renderType=renderType) {
        sink_hose_extender();
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

