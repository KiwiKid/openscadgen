

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
    partType = "main";

    holderSlide = 10;
    holderWidth = 50;
    holderHeight = 5;
    
    holderCubeHeight = 20;
    holderCubeWidth = 3;
    holderSize = [100, holderCubeHeight, holderCubeWidth];
    
    
    toggleSlideHeight = 1;

    toggleSlideWidth = 13;
    
    toggleGapWidth = 10;
    
    toggleHeight = 7;
    
    
    connectorSlide = holderCubeWidth;
    
    connectorWidth = 14;
    
    module toggleSlide(){
        up(holderCubeWidth/2)
        cuboid([25, toggleGapWidth,holderCubeWidth]);
       
        cuboid([50,toggleSlideWidth+1,1.5], chamfer=1, edges=TOP);
        
        
//            left(holderCubeWidth)
        left(toggleSlideWidth/2+4)
        
        rotate([90,0,270])
                   dovetail("female", slide=connectorSlide, width=connectorWidth, height=holderHeight);
            
    
    }

	module shopping_list_modula(){
        difference(){
            cuboid(holderSize, rounding=0.1);
            
            back(holderCubeHeight/2)
            rotate([270,0,0])
            dovetail("female", slide=holderSlide, width=holderWidth, height=holderHeight);
            
            fwd(2)
            left(40)
            
            toggleSlide();

            fwd(4)
            right(10)
            text3d("Shopping list item", h=10, size=6, anchor=BOTTOM);
        }
        
        fwd(holderCubeHeight/2)
        rotate([90,0,0])
        dovetail("male", slide=holderHeight/2, width=holderWidth, height=holderHeight);

	}


    sliced(renderType=renderType) {
        if(partType == "main"|| partType == "all"){
            shopping_list_modula();
        } 
       
       if(partType == "connector" || partType == "all"){
       rotate([90,0,90])
            dovetail("male", slide=connectorSlide, width=connectorWidth, height=holderHeight);
        } 
       
       left(30)
       if(partType == "toggle" || partType == "all"){
       up(toggleHeight/2)
           cyl(d=toggleGapWidth, h=toggleHeight);
           right(4)
           cuboid([20, toggleSlideWidth-1 , toggleSlideHeight], chamfer=1, edges=TOP);
        }
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

