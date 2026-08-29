

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
    partType = "clip"; //all | base clip

    remoteHolderHeight = 13;
    
    remoteControlWidth = 52;
    remoteControlDepth = 26;
    remoteSize = [remoteHolderHeight+1, remoteControlDepth, remoteControlWidth];
    
       cutoutWidth = 7;
   remoteHolderOuterSize = [remoteHolderHeight, remoteSize[1]+cutoutWidth, remoteSize[2]+cutoutWidth];
    
    //remoteCutoutSize = [remoteHolderHeight+1,remoteSize[1]-cutoutWidth,remoteSize[2]-cutoutWidth];
    
    
    remoteCutoutSize2 = [remoteHolderHeight+1,24,38];
    
    screwOffset = 10;
    
    module screwHoles(){
        back(20)
      down(screwOffset) 
      
      rotate([0,90,90])
        cyl(12, 2, chamfer1=-2);
       
        back(20)
      up(screwOffset) 
      rotate([0,90,90])
       cyl(12, 2, chamfer1=-2);
    }

	module remote_clip(){
    difference(){

		cuboid(remoteHolderOuterSize, rounding=2);
        
        cuboid(remoteSize);
        fwd(20)
        cuboid(remoteCutoutSize2);
        
        back(-cutoutWidth+2)
        rotate([0,0,0])
       screwHoles();
        }
	}

    module clip_base(){
        baseSize= [3,32,26];
        holderSize = [12,baseSize[1],10];
        
       
       cuboid(baseSize, anchor=RIGHT, rounding=1);
        up(holderSize[1]/2)
        left(holderSize[0]/2)
       difference(){
       cuboid(holderSize, rounding=1);
       down(19.1)
       left(1)
       rotate([90,0,0])
       screwHoles();
       }
    }

    sliced(renderType=renderType) {
    if(partType == "clip" || partType == "all"){

        remote_clip();
        
        }
        if(partType =="base"  || partType == "all"){
        
        clip_base();
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

