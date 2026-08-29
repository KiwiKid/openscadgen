

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
    
    // [[-20,70], [100,70]]
    /*
  //  textLineOne = "WE STAND";
  
     // [[10,10], [110,10]],
    textLineTwo = "WITH OUR";
    
    textLineThree = "NURSES";
    */
    
      textLineOne = "VOTE FOR";
      textLineTwo = "MORE BIKE";
    textLineThree = "LANES";
    
//    textLineOne = "VOTE";
    
     
    
     //textLineTwo = "NOW";
    
   // textLineThree = "BIKES";
    
    //supports = [[[20,80], [110,80]], [[10,10], [110,10]], [[40,-60], [110,-60]], [[110,-60], [-50,-60]]] ; 
    
    
    textHeight = 60;
    textOffset = 6;
    plateSize = 225;
    plateDepth = 1;
    
    width=10;
    
    module standard_width_text( height=10, desired_size = [255, 90,5], txt = "Hello"){
       resize(desired_size)
       linear_extrude(height)
       text(txt, anchor=center, font = "FontName:style=Bold");
    }
    
    
    module support_bar(path=path, width=width, depth=depth){
        linear_extrude(depth)
        stroke(path, width=width);
    }

	module text_sign(plateSize=plateSize, plateDepth=plateDepth, textHeight=textHeight, textOffset=textOffset){
    difference(){
   // back(textHeight/2)
    //right(plateSize/2)
        up(2)
		cuboid([plateSize,plateSize,plateDepth], rounding=40, edges="Z");
        
        
        desired_size = [plateSize*0.86, textHeight,5];
        
            standard_width_text(txt=textLineTwo, desired_size=desired_size);
            fwd(textHeight+textOffset)
             standard_width_text(txt=textLineThree,desired_size=desired_size);
            back(textHeight+textOffset)
              standard_width_text(txt=textLineOne, desired_size=desired_size);
        
        }
  
        
        if(textLineThree == "NURSES") {
            supports = [[[20,70], [100,70]],  [[10,10], [110,10]], [[5,-55], [-30,-55]]];

                   for (a = [ 0 : len(supports) - 1 ]){

                        point=supports[a];
                        up(1.5)
                          support_bar(path=point, width=plateDepth, depth=plateDepth);
                  }
              } else if(textLineThree == "DOCTORS") {
                    supports = [ [[20,70], [100,70]], [[10,10], [110,10]], [[-100,-65], [-40,-65]], [[45,-55], [70,-55]], [[10,-65], [45,-65]]];

                   for (a = [ 0 : len(supports) - 1 ]){

                        point=supports[a];
                        up(1.5)
                          support_bar(path=point, width=plateDepth, depth=plateDepth);
                  }
              
              } else if(textLineThree == "TEACHERS") {
                    supports = [ [[20,70], [100,70]], [[10,10], [110,10]], [[-50,-65], [-30,-65]], [[40,-55], [75,-55]]];

                   for (a = [ 0 : len(supports) - 1 ]){

                        point=supports[a];
                        up(1.5)
                          support_bar(path=point, width=plateDepth, depth=plateDepth);
                  }
                } else if(textLineThree == "(NOT PHIL)"){
                    supports = [ [[-50,60], [10,60]], [[-40,0], [110,0]], [[-60,-60], [-30,-60]], [[0,-53], [28,-53]]];

                       for (a = [ 0 : len(supports) - 1 ]){

                        point=supports[a];
                        up(1.5)
                          support_bar(path=point, width=plateDepth, depth=plateDepth);
                  }
                
                } else if(textLineThree == "LANES"){
                supports = [ [[-100,60], [-40,60]], [[100,80], [40,60]], [[-70,13], [110,13]], [[-70,-13], [110,-13]], [[-60,-60], [-30,-60]]];

                       for (a = [ 0 : len(supports) - 1 ]){

                        point=supports[a];
                        up(1.5)
                          support_bar(path=point, width=plateDepth, depth=plateDepth);
                  }
                
                }
	}


    sliced(renderType=renderType) {
        text_sign();
    }
       
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.1,
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




	
   
